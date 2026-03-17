package analyzer

import (
	"math"
	"time"

	"github.com/asiffer/netspot/collector"
	"github.com/asiffer/netspot/register"

	"github.com/asiffer/gospot"
)

var NaN = math.NaN()

const OPERATION = "op"

const MIN_DATA = 20

const (
	NORMAL     = int(gospot.NORMAL)
	EXCESS     = int(gospot.EXCESS)
	ANOMALY    = int(gospot.ANOMALY)
	RETURN_NAN = -1
)

type Stat interface {
	// Name returns the unique name of the statistics
	Name() string
	// Update must return the compuetd stat or NaN
	Update(data *collector.Data) float64
}

// MonitoredStat is a wrapper around a Stat that uses the SPOT algorithm
// to detect anomalies
type MonitoredStat struct {
	stat             Stat
	spot             *gospot.Spot
	trainingSet      []float64
	trainingSetIndex int
}

func Monitor(s Stat, options ...SpotOption) *MonitoredStat {
	spot := NewSpot(options...)
	m := &MonitoredStat{
		stat:             s,
		spot:             spot,
		trainingSetIndex: 0,
	}
	// if level is 0.98 and minData is 20, it means that the 0.02 highest values represent
	// about 20 values. So we need 20/(1-0.98) = 1000 training data to reach this minData.
	minSize := int(float64(MIN_DATA) / (1 - m.spot.Level))
	return m.WithTrainingSetSize(minSize)
}

// SetTrainingSetSize sets the size of the training set
func (m *MonitoredStat) WithTrainingSetSize(size int) *MonitoredStat {
	m.trainingSet = make([]float64, size)
	m.trainingSetIndex = 0
	return m
}

func (m *MonitoredStat) Trained() bool {
	return m.trainingSet == nil
}

// Flag computes the stat and analyzes it through the SPOT algorithm
func (m *MonitoredStat) Flag(data *collector.Data) (float64, int, error) {
	value := m.stat.Update(data)
	if math.IsNaN(value) {
		return value, RETURN_NAN, nil
	}
	if m.Trained() {
		// TODO: need to dispatch anomalies
		return value, int(m.spot.Step(value)), nil
	} else {
		m.trainingSet[m.trainingSetIndex] = value
		m.trainingSetIndex++

		// last training data
		if m.trainingSetIndex == len(m.trainingSet) {
			if err := m.spot.Fit(m.trainingSet); err != nil {
				// rollback
				m.trainingSetIndex--
				return value, NORMAL, err
			}
			// we don't need this training set anymore
			m.trainingSet = nil
			m.trainingSetIndex = -1
		}
	}
	return value, NORMAL, nil
}

type Alert struct {
	Probability float64     `json:"probability,format:nonfinite"`
	State       gospot.Spot `json:"-"`
}

type StatValue struct {
	Value            float64 `json:"value,format:nonfinite"`             // stat value
	ExcessThreshold  float64 `json:"excess_threshold,format:nonfinite"`  // Excess threshold
	AnomalyThreshold float64 `json:"anomaly_threshold,format:nonfinite"` // Anomaly threshold
	Alert            *Alert  `json:"alert,omitempty"`                    // Alert if the value is an anomaly
}

type Record struct {
	Time   time.Time             `json:"time"`
	Stats  map[string]*StatValue `json:"stats"`
	Alerts []string              `json:"alerts,omitempty"`
}

func NewRecord() Record {
	return Record{
		Stats: make(map[string]*StatValue),
	}
}

// MonitoredStatsList stores all the monitored statistics
type MonitoredStatsList struct {
	monitoredStats []*MonitoredStat
	alertHooks     *register.Register2[string, *StatValue]
	valueHooks     *register.Register[*Record]
	spotErrorHooks *register.Register2[string, *SpotError]
}

func NewMonitoredStatsList() *MonitoredStatsList {
	return &MonitoredStatsList{
		monitoredStats: make([]*MonitoredStat, 0),
		alertHooks:     register.NewRegister2[string, *StatValue](),
		valueHooks:     register.NewRegister[*Record](),
		spotErrorHooks: register.NewRegister2[string, *SpotError](),
	}
}

// Add a new monitored stat to the list
func (m *MonitoredStatsList) Add(ms *MonitoredStat) {
	m.monitoredStats = append(m.monitoredStats, ms)
}

// OnAlert registers a new hook to be called when an alert is triggered
func (m *MonitoredStatsList) OnAlert(hook func(string, *StatValue)) *MonitoredStatsList {
	m.alertHooks.Register(hook)
	return m
}

// OnValue registers a new hook to be called when stats are updated
func (m *MonitoredStatsList) OnData(hook func(*Record)) *MonitoredStatsList {
	m.valueHooks.Register(hook)
	return m
}

func (m *MonitoredStatsList) OnSpotError(hook func(string, *SpotError)) *MonitoredStatsList {
	m.spotErrorHooks.Register(hook)
	return m
}

// Hook is a method aimed to be passed to the chosen collector
func (m *MonitoredStatsList) Hook(data *collector.Data) {
	record := NewRecord()
	record.Time = time.Unix(0, int64(data.TIME))
	for _, s := range m.monitoredStats {
		value, result, err := s.Flag(data)
		if err != nil {
			m.spotErrorHooks.Exec(s.stat.Name(), &SpotError{
				Err:   err,
				State: s.spot,
			})
			continue
		}
		// populate stat values to dispatch to hooks
		sv := StatValue{
			Value:            value,
			ExcessThreshold:  s.spot.ExcessThreshold,
			AnomalyThreshold: s.spot.AnomalyThreshold,
		}
		if result == ANOMALY {
			sv.Alert = &Alert{
				Probability: s.spot.Probability(value),
				State:       *s.spot,
			}
			record.Alerts = append(record.Alerts, s.stat.Name())
			// also dispatch the alert to alert hooks
			m.alertHooks.Exec(s.stat.Name(), &sv)
		}
		record.Stats[s.stat.Name()] = &sv

	}
	m.valueHooks.Exec(&record)
}
