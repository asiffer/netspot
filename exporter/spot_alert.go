// spot_alert.go
//

package exporter

// SpotAlert is a simple structure to log alerts sent
// by spot instances
type SpotAlert struct {
	Status      string
	Stat        string
	Value       float64
	Code        int
	Probability float64
}

func (s *SpotAlert) toUntypedMap() map[string]interface{} {
	return map[string]interface{}{
		"status":      s.Status,
		"stat":        s.Stat,
		"value":       s.Value,
		"code":        s.Code,
		"probability": s.Probability,
	}
}
