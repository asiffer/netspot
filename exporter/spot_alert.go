// spot_alert.go
//

package exporter

import (
	"fmt"
	"time"
)

const SpotAlertJsonFormat = "\"status\":\"%s\",\"stat\":\"%s\",\"value\":%f,\"code\":%d,\"probability\":%f"

// SpotAlert is a simple structure to log alerts sent
// by spot instances
type SpotAlert struct {
	Status      string  // UP_ALERT or DOWN_ALERT
	Stat        string  // Stat is the name of the statistic
	Value       float64 // Value is the [abnormal] value if the statistic
	Code        int     // Code is a return code of the SPOT algorithms (useless here)
	Probability float64 // Probability corresponds to the probability to see an event at least as extreme (higher or lower depending on the Status)
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

func (s *SpotAlert) toJSON() string {
	return fmt.Sprintf("{%s}",
		fmt.Sprintf(SpotAlertJsonFormat,
			s.Status,
			s.Stat,
			s.Value,
			s.Code,
			s.Probability,
		),
	)
}

func (s *SpotAlert) toJSONwithTime(t time.Time) string {
	return fmt.Sprintf("{\"time\":%d,%s}",
		t.UnixNano(),
		fmt.Sprintf(SpotAlertJsonFormat,
			s.Status,
			s.Stat,
			s.Value,
			s.Code,
			s.Probability,
		),
	)
}
