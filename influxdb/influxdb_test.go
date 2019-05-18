// influxdb_test.go

package influxdb

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var (
	headerWidth    = 80
	configTestFile = "test/influxdb.toml"
)

func checkTitle(s string) {
	format := "%-" + fmt.Sprint(headerWidth-7) + "s"
	fmt.Printf(format, s)
}

func testOK() {
	fmt.Println("[\033[32mOK\033[0m]")
}

func testERROR() {
	fmt.Println("[\033[31mERROR\033[0m]")
}

func init() {
	zerolog.SetGlobalLevel(zerolog.NoLevel)
}

func title(s string) {
	var l = len(s)
	var border int
	var left string
	var right string
	remaining := headerWidth - l - 2
	if remaining%2 == 0 {
		border = remaining / 2
		left = strings.Repeat("-", border) + " "
		right = " " + strings.Repeat("-", border)
	} else {
		border = (remaining - 1) / 2
		left = strings.Repeat("-", border+1) + " "
		right = " " + strings.Repeat("-", border)
	}

	fmt.Println(left + s + right)

}

func TestInitConfig(t *testing.T) {
	title("Testing config file loading")
	viper.SetConfigFile(configTestFile)
	viper.ReadInConfig()
	InitConfig()

	checkTitle("Checking address...")
	if GetAddress() != "http://127.0.0.1:8086" {
		testERROR()
		t.Errorf("Expected 'http://127.0.0.1:8086', got '%s'", GetAddress())
	} else {
		testOK()
	}

	checkTitle("Checking username...")
	if username != "John" {
		testERROR()
		t.Errorf("Expected 'John', got '%s'", username)
	} else {
		testOK()
	}

	checkTitle("Checking password...")
	if password != "Doe" {
		testERROR()
		t.Errorf("Expected 'Doe', got '%s'", password)
	} else {
		testOK()
	}

	checkTitle("Checking database...")
	if database != "db" {
		testERROR()
		t.Errorf("Expected 'db', got '%s'", database)
	} else {
		testOK()
	}

	checkTitle("Checking batch size...")
	if batchSize != 23 {
		testERROR()
		t.Errorf("Expected 23, got %d", batchSize)
	} else {
		testOK()
	}

	checkTitle("Checking agent name...")
	if agentName != "007" {
		testERROR()
		t.Errorf("Expected '007', got '%s'", agentName)
	} else {
		testOK()
	}

}

func TestPushRecord(t *testing.T) {
	title("Testing push records")
	PushRecord(map[string]float64{"a": 12., "b": -3.5},
		"seriesName", time.Unix(0, 0))
	checkTitle("Checking ID...")
	if batchID != 1 {
		testERROR()
		t.Errorf("Expected 1, got %d", batchID)
	} else {
		testOK()
	}

	p := batch.Points()[0]
	checkTitle("Checking record time...")
	if p.Time() != time.Unix(0, 0) {
		testERROR()
		t.Errorf("Expected %s, got %s", time.Unix(0, 0), p.Time())
	} else {
		testOK()
	}

	checkTitle("Checking series name...")
	if p.Name() != "seriesName" {
		testERROR()
		t.Errorf("Expected %s, got %s", "seriesName", p.Name())
	} else {
		testOK()
	}

	checkTitle("Checking agent name...")
	if p.Tags()["agent"] != agentName {
		testERROR()
		t.Errorf("Expected %s, got %s", agentName, p.Tags()["agent"])
	} else {
		testOK()
	}

	checkTitle("Checking values...")
	fields, err := p.Fields()
	if err != nil {
		testERROR()
		t.Error(err.Error())
	} else {
		a := fields["a"].(float64)
		b := fields["b"].(float64)
		if a != 12. || b != -3.5 {
			testERROR()
			t.Errorf("Expected a=%f and b=%f, got a = %f and b = %f", 12., -3.5, a, b)
		} else {
			testOK()
		}
	}

}
