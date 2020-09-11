package exporter

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func sortedKeys(m map[string]float64) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	sort.Strings(keys)
	return keys
}

func jsonify(data map[string]float64) string {
	out := "{%s}"
	out2 := make([]string, len(data))

	for i, key := range sortedKeys(data) {
		out2[i] = fmt.Sprintf("\"%s\":%f", key, data[key])
	}
	return fmt.Sprintf(out, strings.Join(out2, ","))
}

func jsonifyWithTime(t time.Time, data map[string]float64) string {
	out := "{%s}"
	out2 := make([]string, len(data)+1) // + time
	out2[0] = fmt.Sprintf("\"%s\":%d", "time", t.UnixNano())

	for i, key := range sortedKeys(data) {
		out2[i+1] = fmt.Sprintf("\"%s\":%f", key, data[key])
	}
	return fmt.Sprintf(out, strings.Join(out2, ","))
}
