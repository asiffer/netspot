// analyzer.go
package analyzer

// import "netspot/initconfig"

import (
	"errors"
	"fmt"
	"gospot"
	"math"
	"netspot/influxdb"
	"netspot/miner"
	"netspot/stats"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//------------------------------------------------------------------------------
// GLOBAL VARIABLES
//------------------------------------------------------------------------------

var (
	counterId     map[string]int              // map CounterName -> counterId (within miner CounterMap)
	statMap       map[int]stats.StatInterface // map StatId -> Stat
	statId        int                         // id of the next loaded stat
	statValues    map[string]float64          // the last computed values of the statistics
	counterValues map[string]uint64           // temp container of the counter values
	period        time.Duration               // time between two stat updates (= window size)
	mux           sync.RWMutex                // Locker for the counter map access
	events        chan int                    // channel to sen/receive events
	running       bool                        // to check if the stat are computed
)

var (
	logDataToFile       bool           // if data/thresholds are logged to file
	logDataToInfluxDB   bool           // if data/thresholds are logged to influxdb
	rawDataLogger       zerolog.Logger // log raw statistics
	thresholdLogger     zerolog.Logger // log raw statistics thresholds
	rawDataOutputFile   string         // the path of the file containing data
	thresholdOutputFile string         // the path of the file containing thresholds
	seriesName          string         // the name of the influxdb series
)

var err error

//------------------------------------------------------------------------------
// INITIALIZATION
//------------------------------------------------------------------------------

// init sets the default configuration
func init() {
	// default config
	viper.SetDefault("analyzer.period", 2*time.Second)
	viper.SetDefault("analyzer.datalog.influxdb", false)
	viper.SetDefault("analyzer.datalog.file", true)
	viper.SetDefault("analyzer.datalog.output_dir", "/tmp")
	viper.SetDefault("analyzer.stats", []string{})
}

// init initializes package global variables
func init() {
	counterId = make(map[string]int)
	statMap = make(map[int]stats.StatInterface)
	statId = 0
	statValues = make(map[string]float64)
	counterValues = make(map[string]uint64)
	events = make(chan int)
	running = false
}

// func init() {
func InitConfig() {
	// settings
	SetPeriod(viper.GetDuration("analyzer.period"))
	logDataToFile = viper.GetBool("analyzer.datalog.file")
	logDataToInfluxDB = viper.GetBool("analyzer.datalog.influxdb")
	// SetSeriesName(seriesNameFromCurrentTime())

	if logDataToInfluxDB {
		influxdb.InitConfig()
	}

	// if logDataToFile {
	// 	p := path.Join(viper.GetString("analyzer.datalog.output_dir"), "nsraw_"+seriesName+".json")
	// 	f, err := os.Create(p)
	// 	if err != nil {
	// 		log.Fatal().Msg("Error while creating raw data log file")
	// 	}
	// 	rawDataLogger = zerolog.New(f).With().Logger()
	// 	rawDataOutputFile = f.Name()
	// 	log.Debug().Msgf("Raw statistics saved to %s", rawDataOutputFile)

	// 	p = path.Join(viper.GetString("analyzer.datalog.output_dir"), "nsthreshold_"+seriesName+".json")
	// 	f, err = os.Create(p)
	// 	if err != nil {
	// 		log.Fatal().Msg("Error while creating threshold log file")
	// 	}
	// 	thresholdLogger = zerolog.New(f).With().Logger()
	// 	thresholdOutputFile = f.Name()
	// 	log.Debug().Msgf("Thresholds saved to %s", thresholdOutputFile)

	// } else {
	// 	rawDataLogger = zerolog.New(nil).With().Logger()
	// 	thresholdLogger = zerolog.New(nil).With().Logger()
	// }

	for _, s := range viper.GetStringSlice("analyzer.stats") {
		LoadFromName(s)
	}

	log.Debug().Msg(fmt.Sprint("Available stats: ", GetAvailableStats()))
	log.Info().Msg("Analyzer package configured")

}

//------------------------------------------------------------------------------
// UNEXPORTED FUNCTIONS
//------------------------------------------------------------------------------

// GENERIC ---------------------------------------------------------------------

func find(sl []string, str string) int {
	for i, s := range sl {
		if s == str {
			return i
		}
	}
	return -1
}

// SPECIFIC --------------------------------------------------------------------

func seriesNameFromCurrentTime() string {
	t := strings.Replace(time.Now().Format(time.Stamp), "_", "0", -1)
	t = strings.Replace(t, " ", "_", -1)
	return t
}

// func release() {
// 	running = false
// 	// if logDataToInfluxDB {
// 	// 	influxdb.Close()
// 	// }
// }

func release() {
	running = false
	close(events)
}

func updateAndReset() {
	mux.Lock()
	var val uint64
	var err error
	for ctrname, ctrid := range counterId {
		val, err = miner.GetCounterValue(ctrid)
		if err != nil {
			log.Error().Msgf("The ID of %s (%d) is wrong for the miner", ctrname, ctrid)
		} else {
			counterValues[ctrname] = val
			miner.Reset(ctrid)
		}
	}
	log.Debug().Msg(fmt.Sprint(counterValues))
	mux.Unlock()
}

func isLoaded(statname string) int {
	for i, stat := range statMap {
		if statname == stat.Name() {
			return i
		}
	}
	return -1
}

func getcounterValues(ctrnames []string) []uint64 {
	values := make([]uint64, len(ctrnames))
	for i, name := range ctrnames {
		values[i] = counterValues[name]
	}
	return values
}

func compute(id int) float64 {
	counters := statMap[id].Requirement()
	values := getcounterValues(counters)
	return statMap[id].Compute(values)
}

func load(stat stats.StatInterface) (int, error) {
	var id int
	if stat != nil {
		if isLoaded(stat.Name()) > 0 {
			log.Debug().Msgf("Stat %s already loaded", stat.Name())
			msg := fmt.Sprintf("Stat %s already loaded", stat.Name())
			return -2, errors.New(msg)
		} else {
			// load the counters
			for _, ctrname := range stat.Requirement() {
				id = miner.LoadFromName(ctrname)
				if id > 0 {
					counterId[ctrname] = id
				}
			}
			// increment the stat container
			statId = statId + 1
			statMap[statId] = stat
			log.Debug().Msgf("Loading stat %s", stat.Name())
			return statId, nil
		}
	}
	return -1, errors.New("Cannot load null stat")
}

// unload
func unload(id int) (int, error) {
	var index int
	// check the potential counters to remove
	stat, exists := statMap[id]
	if !exists {
		msg := fmt.Sprintf("Unknown Stat id %d", id)
		return -1, errors.New(msg)
	}
	counters2remove := stat.Requirement()
	// look if these counters are requested by other stats
	for i, stat := range statMap {
		if i != id {
			for _, req := range stat.Requirement() {
				index = find(counters2remove, req)
				if index >= 0 {
					// if a stat still need this counter, we remove it from the list
					counters2remove = append(counters2remove[:index], counters2remove[(index+1):]...)
				}
			}
		}
	}
	// We remove all the useless counters
	for _, ctr := range counters2remove {
		delete(counterId, ctr)
		delete(counterValues, ctr)
		miner.UnloadFromName(ctr)
	}
	// we remove the stat
	delete(statMap, id)
	return 0, nil
}

// checkSpotOutput analyze the output provided by (D)SPOT.
// id is the stat identifier in statMap
// val is the stat value
// res is the (D)Spot output
func checkSpotOutput(id int, val float64, res int32) {
	if res == 1 {
		log.Warn().Str("Status", "UP_ALERT").
			Str("Stat", statMap[id].Name()).
			Float64("Value", val).Int32("Spot", res).
			Float64("Probability", statMap[id].DSpot().UpProbability(val)).Msg("Alarm!")
	} else if res == -1 {
		log.Warn().Str("Status", "DOWN_ALERT").
			Str("Stat", statMap[id].Name()).
			Float64("Value", val).Int32("Spot", res).
			Float64("Probability", statMap[id].DSpot().DownProbability(val)).Msg("Alarm!")
	}
}

//------------------------------------------------------------------------------
// EXPORTED FUNCTIONS
//------------------------------------------------------------------------------

// StatStatus returns the status of the dspot instance monitoring that stat
func StatStatus(s string) (gospot.DSpotStatus, error) {
	i := isLoaded(s)
	if i >= 0 {
		return statMap[i].DSpot().Status(), nil
	}
	msg := fmt.Sprintf("%s statistics is not loaded", s)
	return gospot.DSpotStatus{}, errors.New(msg)
}

// RawStatus returns the current status of the analyzer through a
// basic map. It is designed to a future print.
func RawStatus() map[string]string {
	m := make(map[string]string)
	m["period"] = fmt.Sprint(period)
	if logDataToInfluxDB {
		m["influxdb logging"] = fmt.Sprintf("%v (%s)", logDataToInfluxDB, influxdb.GetAddress())
	} else {
		m["influxdb logging"] = "no"
	}

	if logDataToFile {
		m["file logging"] = fmt.Sprintf("%v (%s)", logDataToFile, viper.GetString("analyzer.datalog.output_dir"))
	} else {
		m["file logging"] = "no"
	}
	m["statistics"] = fmt.Sprint(GetLoadedStats())
	return m
}

// InitDataLogging creates new loggers (file/influxdb). Normally, it is called when
// start to run.
func InitDataLogging() {
	SetSeriesName(seriesNameFromCurrentTime())

	if logDataToFile {
		p := path.Join(viper.GetString("analyzer.datalog.output_dir"), "netspot_raw_"+seriesName+".json")
		f, err := os.Create(p)
		if err != nil {
			log.Fatal().Msg("Error while creating raw data log file")
		}
		rawDataLogger = zerolog.New(f).With().Logger()
		rawDataOutputFile = f.Name()

		p = path.Join(viper.GetString("analyzer.datalog.output_dir"), "netspot_threshold_"+seriesName+".json")
		f, err = os.Create(p)
		if err != nil {
			log.Fatal().Msg("Error while creating threshold log file")
		}
		thresholdLogger = zerolog.New(f).With().Logger()
		thresholdOutputFile = f.Name()
	} else {
		rawDataLogger = zerolog.New(nil).With().Logger()
		thresholdLogger = zerolog.New(nil).With().Logger()
	}
}

// Zero aims to zero the internal state of the miner. So it removes all
// the loaded stats, initialize some variables and read the config file.
// In particular it creates a new series name, so the data logs will be placed in
// new files/db.series
func Zero() error {
	if !IsRunning() {
		SetPeriod(viper.GetDuration("analyzer.period"))
		// period = viper.GetDuration("analyzer.period")
		logDataToFile = viper.GetBool("analyzer.datalog.file")
		logDataToInfluxDB = viper.GetBool("analyzer.datalog.influxdb")
		// SetSeriesName(seriesNameFromCurrentTime())

		// if logDataToFile {
		// 	p := path.Join(viper.GetString("analyzer.datalog.output_dir"), "nsraw_"+seriesName+".json")
		// 	f, err := os.Create(p)
		// 	if err != nil {
		// 		log.Fatal().Msg("Error while creating raw data log file")
		// 	}
		// 	rawDataLogger = zerolog.New(f).With().Logger()
		// 	rawDataOutputFile = f.Name()

		// 	p = path.Join(viper.GetString("analyzer.datalog.output_dir"), "nsthreshold_"+seriesName+".json")
		// 	f, err = os.Create(p)
		// 	if err != nil {
		// 		log.Fatal().Msg("Error while creating threshold log file")
		// 	}
		// 	thresholdLogger = zerolog.New(f).With().Logger()
		// 	thresholdOutputFile = f.Name()
		// } else {
		// 	rawDataLogger = zerolog.New(nil).With().Logger()
		// 	thresholdLogger = zerolog.New(nil).With().Logger()
		// }

		// package variables
		counterId = make(map[string]int)
		statMap = make(map[int]stats.StatInterface)
		statId = 0
		statValues = make(map[string]float64)
		counterValues = make(map[string]uint64)
		// stopChan = make(chan int)
		events = make(chan int)
		running = false

		log.Info().Msg("Analyzer package reloaded")
		return nil
	} else {
		log.Error().Msg("Cannot reload, monitoring in progress")
		return errors.New("Cannot reload, monitoring in progress")
	}
}

// Config return the configuration of the analyzer/miner
func Config() map[string]interface{} {
	return viper.AllSettings()
}

// DisableLogging disable the log output. Warning! It disables the log
// for all the modules using zerolog
func DisableLogging() {
	log.Info().Msg("Disabling logging")
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// SetLogging set the minimum level of the output logs.
// - panic (zerolog.PanicLevel, 5)
// - fatal (zerolog.FatalLevel, 4)
// - error (zerolog.ErrorLevel, 3)
// - warn (zerolog.WarnLevel, 2)
// - info (zerolog.InfoLevel, 1)
// - debug (zerolog.DebugLevel, 0)
func SetLogging(level int) {
	l := zerolog.Level(level)
	zerolog.SetGlobalLevel(l)
	log.Info().Msgf("Enabling logging (level %s)", l.String())
}

// SetPeriod sets the duration between two stat computations
func SetPeriod(d time.Duration) {
	period = d
	log.Debug().Msgf("Period set to %s", d)
}

// SetSeriesName sets the name of the series within InfluxDB
func SetSeriesName(s string) {
	seriesName = s
	log.Debug().Msgf(`Series name set to "%s"`, s)
}

// GetPeriod returns the current duration between two stat computations
func GetPeriod() time.Duration {
	return period
}

// GetLoadedStats returns the slice of the names of the loaded statistics
func GetLoadedStats() []string {
	list := make([]string, 0)
	for _, s := range statMap {
		list = append(list, s.Name())
	}
	return list
}

// GetNumberOfLoadedStats returns the number of loaded statistics
func GetNumberOfLoadedStats() int {
	return len(statMap)
}

// GetAvailableStats returns the slice of the names of the available
// statistics
func GetAvailableStats() []string {
	list := make([]string, 0)
	for name, _ := range stats.AVAILABLE_STATS {
		list = append(list, name)
	}
	return list
}

// LoadFromName loads the statistics corresponding to the given name
// and returns the id where it is internally stored. An error is returned
// when the statistics is unknown.
func LoadFromName(statname string) (int, error) {
	stat, err := stats.StatFromName(statname)
	if err != nil {
		return -1, err
	}
	id, err := load(stat)
	return id, err
}

// UnloadFromName removes the statistics, so it will not be monitored.
// It returns 0, nil if the unload is ok, or -1, error otherwise.
func UnloadFromName(statname string) (int, error) {
	id := isLoaded(statname)
	if id > 0 {
		ret, err := unload(id)
		return ret, err
	} else {
		msg := fmt.Sprintf("%s statistics is not loaded", statname)
		return -1, errors.New(msg)
	}
}

// UnloadAll removes all the previously loaded statistics
func UnloadAll() {
	for i, _ := range statMap {
		delete(statMap, i)
	}
	for i, _ := range counterValues {
		delete(counterValues, i)
	}
	for i, _ := range counterId {
		delete(counterId, i)
	}
	miner.UnloadAll()
	statId = 0
}

// IsRunning checks whether the statistics are currently computed
func IsRunning() bool {
	return running
}

// StopStats stops the analysis
func StopStats() {
	if running {
		events <- 0
		// log.Info().Msg("Stopping stats computation")
	}
}

// StartStats starts the analysis
// func StartStats() error {
func StartStats() error {
	if miner.IsSniffing() {
		log.Info().Msg("Starting stats computation")
		log.Debug().Msg(fmt.Sprint("Loaded stats: ", GetLoadedStats()))
		events = make(chan int)
		go run()
	} else {
		log.Error().Msg("The counters have not been launched")
		return errors.New("The counters have not been launched")
	}
	return nil
}

// StartStatsAndWait starts the analysis. It will stop only when no packets
// have to be processed (ex: pcap file)
func StartStatsAndWait() error {
	if miner.IsSniffing() {
		log.Info().Msg("Starting stats computation")
		log.Debug().Msg(fmt.Sprint("Loaded stats: ", GetLoadedStats()))
		events = make(chan int)
		run()
	} else {
		log.Error().Msg("The counters have not been launched")
		return errors.New("The counters have not been launched")
	}
	return nil
}

//------------------------------------------------------------------------------
// PARAMOUNT BUT UNEXPORTED FUNCTIONS
//------------------------------------------------------------------------------

// analyze loops over the computed stats and send it to DSpot instances
// The stat values and the computed thresholds are logged (file and/or influxdb)
func analyze() {
	var val float64
	var up_th, down_th float64
	var res int32
	var name string
	curtime := miner.SourceTime

	dlog := rawDataLogger.Log().Time("time", curtime)
	tlog := thresholdLogger.Log().Time("time", curtime)

	for id, stat := range statMap {
		name = stat.Name()

		// log thresholds
		up_th = stat.DSpot().GetUpperThreshold()
		down_th = stat.DSpot().GetLowerThreshold()

		// if up_th is NaN, it means that up data are not monitored or
		// the calibration has not finished
		if !math.IsNaN(up_th) {
			tlog.Float64(name+"_UP", up_th)
			statValues[name+"_UP"] = up_th
		}

		// if down_th is NaN, it means that down data are not monitored or
		// the calibration has not finished
		if !math.IsNaN(down_th) {
			tlog.Float64(name+"_DOWN", down_th)
			statValues[name+"_DOWN"] = down_th
		}

		// compute the statistics
		val = stat.Compute(getcounterValues(stat.Requirement()))
		// feed DSpot
		res = stat.Update(val)
		// check alert
		checkSpotOutput(id, val, res)

		// log data
		dlog.Float64(name, val)
		statValues[name] = val

	}
	dlog.Msg("")
	tlog.Msg("")
	// if data have to be sent to InfluxDB
	if logDataToInfluxDB {
		influxdb.PushRecord(statValues, seriesName, curtime)
	}
}

// run open the device to listen
func run() {
	var ticker <-chan time.Time
	var e int

	// initialize files/influxdb to log data and thresholds
	InitDataLogging()

	// Define the clock
	if miner.IsDeviceInterface() {
		seriesName = miner.GetDevice() + "_" + seriesNameFromCurrentTime()
		// live ticker
		ticker = time.Tick(period)
	} else {
		// the series name is then the name of the file
		seriesName = path.Base(miner.GetDevice())
		// the timestamps of the capture define the clock
		ticker = miner.Tick(period)
	}

	// set the running state
	running = true

	// loop
	for {
		select {
		case e = <-events:
			if e == 0 { // stop order
				release()
				log.Info().Msg("Stopping stats computation (controller)")
				return
			}
		case <-ticker:
			if !miner.IsSniffing() {
				release()
				log.Warn().Msg("Stopping stats computation (miner)")
				return
			} else {
				// get the stats values (call miner counters etc.)
				updateAndReset()
				// analyze the stats values (feed dspot, log data/thresholds)
				analyze()
			}
		}
	}
}

//------------------------------------------------------------------------------
// MAIN
//------------------------------------------------------------------------------

func main() {
	// nothing
}
