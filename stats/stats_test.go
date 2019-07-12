// stats_test.go

package stats

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/asiffer/gospot"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var (
	headerWidth         = 80
	configTestFile      = "test/config.toml"
	emptyConfigTestFile = "test/config_empty.toml"
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

func gaussianSample(N int) []float64 {
	rand.Seed(time.Now().UTC().UnixNano())
	data := make([]float64, N)
	for i := 0; i < N; i++ {
		data[i] = rand.NormFloat64()
	}
	return data
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func isEqual(sl1 []string, sl2 []string) bool {
	if len(sl1) != len(sl2) {
		return false
	}
	for _, s := range sl1 {
		if !contains(sl2, s) {
			return false
		}
	}
	return true
}

func TestLoadDSpotConfig(t *testing.T) {
	title("Checking loading DSpot config")
	viper.SetConfigFile(configTestFile)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
	}

	dsc := setCustomConfig("R_SYN")

	checkTitle("Checking depth...")
	if dsc.Depth != 60 {
		testERROR()
		t.Errorf("Expected 60, got %d", dsc.Depth)
	} else {
		testOK()
	}

	checkTitle("Checking q...")
	if dsc.Q != 1e-5 {
		testERROR()
		t.Errorf("Expected 1e-5, got %f", dsc.Q)
	} else {
		testOK()
	}

	checkTitle("Checking n_init...")
	if dsc.Ninit != 1200 {
		testERROR()
		t.Errorf("Expected 1200, got %d", dsc.Ninit)
	} else {
		testOK()
	}

	checkTitle("Checking level...")
	if dsc.Level != 0.999 {
		testERROR()
		t.Errorf("Expected 0.999, got %f", dsc.Level)
	} else {
		testOK()
	}

	checkTitle("Checking up...")
	if !dsc.Up {
		testERROR()
		t.Errorf("Expected true, got %v", dsc.Up)
	} else {
		testOK()
	}

	checkTitle("Checking down...")
	if !dsc.Down {
		testERROR()
		t.Errorf("Expected true, got %v", dsc.Down)
	} else {
		testOK()
	}

	checkTitle("Checking alert...")
	if dsc.Alert {
		testERROR()
		t.Errorf("Expected false, got %v", dsc.Alert)
	} else {
		testOK()
	}

	checkTitle("Checking bounded...")
	if !dsc.Bounded {
		testERROR()
		t.Errorf("Expected true, got %v", dsc.Bounded)
	} else {
		testOK()
	}

	checkTitle("Checking max_excess...")
	if dsc.MaxExcess != 250 {
		testERROR()
		t.Errorf("Expected 250, got %d", dsc.MaxExcess)
	} else {
		testOK()
	}
}

func TestLoadDSpotConfigUnknownStat(t *testing.T) {
	title("Loading default DSpot config")

	dsc := setCustomConfig("WTF")

	checkTitle("Checking depth...")
	if dsc.Depth != 50 {
		testERROR()
		t.Errorf("Expected 50, got %d", dsc.Depth)
	} else {
		testOK()
	}

	checkTitle("Checking q...")
	if dsc.Q != 1e-4 {
		testERROR()
		t.Errorf("Expected 1e-4, got %f", dsc.Q)
	} else {
		testOK()
	}

	checkTitle("Checking n_init...")
	if dsc.Ninit != 1000 {
		testERROR()
		t.Errorf("Expected 1000, got %d", dsc.Ninit)
	} else {
		testOK()
	}

	checkTitle("Checking level...")
	if dsc.Level != 0.98 {
		testERROR()
		t.Errorf("Expected 0.98, got %f", dsc.Level)
	} else {
		testOK()
	}

	checkTitle("Checking up...")
	if !dsc.Up {
		testERROR()
		t.Errorf("Expected true, got %v", dsc.Up)
	} else {
		testOK()
	}

	checkTitle("Checking down...")
	if dsc.Down {
		testERROR()
		t.Errorf("Expected false, got %v", dsc.Down)
	} else {
		testOK()
	}

	checkTitle("Checking alert...")
	if !dsc.Alert {
		testERROR()
		t.Errorf("Expected true, got %v", dsc.Alert)
	} else {
		testOK()
	}

	checkTitle("Checking bounded...")
	if !dsc.Bounded {
		testERROR()
		t.Errorf("Expected true, got %v", dsc.Bounded)
	} else {
		testOK()
	}

	checkTitle("Checking max_excess...")
	if dsc.MaxExcess != 200 {
		testERROR()
		t.Errorf("Expected 200, got %d", dsc.MaxExcess)
	} else {
		testOK()
	}
}

func TestBaseStat(t *testing.T) {
	title("Testing base stat")
	conf := setCustomConfig("R_SYN")
	bs := &BaseStat{
		name:  "Test",
		dspot: gospot.NewDSpotFromConfig(&conf),
	}

	checkTitle("Checking name...")
	if bs.name != "Test" {
		testERROR()
		t.Errorf("Expected 'Test', got %s", bs.name)
	} else {
		testOK()
	}

	checkTitle("Checking update...")
	r := bs.Update(1e-3)
	if r != 5 {
		testERROR()
		t.Errorf("Expected 5, got %d", r)
	} else {
		testOK()
	}

	checkTitle("Getting DSpot instance...")
	if bs.DSpot().Config() != conf {
		testERROR()
		t.Errorf("Bad configuration")
		fmt.Println(bs.DSpot().Config())
	} else {
		testOK()
	}

	checkTitle("Setting custom DSpot config...")
	conf.Ninit = 1500
	bs.SetDSpotConfig(conf)
	if bs.DSpot().Config().Ninit != 1500 {
		testERROR()
		t.Errorf("Expected 1500, got %d", bs.DSpot().Config().Ninit)
	} else {
		testOK()
	}
}

func TestGetStat(t *testing.T) {
	title("Testing getting/registering stats")

	checkTitle("Loading stat...")
	rack, err := StatFromName("R_ACK")
	if rack == nil && err != nil {
		testERROR()
		t.Errorf(err.Error())
	} else {
		testOK()
	}

	checkTitle("Loading unknown stat...")
	wtf, err := StatFromName("WTF")
	if wtf != nil || err == nil {
		testERROR()
		t.Error("The stat WTF should be unknown")
	} else {
		testOK()
	}

	checkTitle("Double registering...")
	err = Register(
		"R_ACK",
		func(bs BaseStat) StatInterface { return nil })
	if err == nil {
		testERROR()
		t.Errorf("R_ACK must be already registered")
	} else {
		testOK()
	}
}

func TestChangeDSpotConfig(t *testing.T) {
	title("Testing DSpot attributes change")
	Ninit := 2000
	MaxExcess := 600
	Up := false
	Down := true
	Alert := false
	Bounded := false
	Level := 0.9999
	Q := 2.3e-8

	extra := map[string]interface{}{
		"q":          Q,
		"n_init":     Ninit,
		"level":      Level,
		"Up":         Up,
		"Down":       Down,
		"Alert":      Alert,
		"bounded":    Bounded,
		"max_excess": MaxExcess,
	}
	rack, _ := StatFromNameWithCustomConfig("R_ACK", extra)

	conf := rack.DSpot().Config()
	if conf.Ninit != Ninit {
		t.Errorf("Expected Ninit equal to %d, got %d", Ninit, conf.Ninit)
	}
	if conf.MaxExcess != MaxExcess {
		t.Errorf("Expected MaxExcess equal to %d, got %d", MaxExcess, conf.MaxExcess)
	}
	if conf.Level != Level {
		t.Errorf("Expected Level equal to %f, got %f", Level, conf.Level)
	}
	if conf.Alert != Alert {
		t.Errorf("Expected Alerts equal to %t, got %t", Alert, conf.Alert)
	}
	if conf.Up != Up {
		t.Errorf("Expected Up equal to %t, got %t", Up, conf.Up)
	}
	if conf.Down != Down {
		t.Errorf("Expected Down equal to %t, got %t", Down, conf.Down)
	}
	if conf.Bounded != Bounded {
		t.Errorf("Expected Bounded equal to %t, got %t", Bounded, conf.Bounded)
	}
	if conf.Q != Q {
		t.Errorf("Expected Q equal to %f, got %f", Q, conf.Q)
	}
}
