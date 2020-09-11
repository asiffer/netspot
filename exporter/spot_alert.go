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
