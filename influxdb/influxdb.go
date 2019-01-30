// influxdb.go
package influxdb

import (
	"fmt"
	"time"

	idb "github.com/influxdata/influxdb/client/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	addr      string
	database  string
	username  string
	password  string
	batchSize int
	batchId   int
	agentName string
)

var (
	client idb.Client
	batch  idb.BatchPoints
	// tag    string
	logger zerolog.Logger
)

func init() {
	// default values
	viper.SetDefault("influxdb.addr", "127.0.0.1")
	viper.SetDefault("influxdb.port", 8086)
	viper.SetDefault("influxdb.database", "netspot")
	viper.SetDefault("influxdb.username", "netspot")
	viper.SetDefault("influxdb.password", "netspot")
	viper.SetDefault("influxdb.batch_size", 10)
	viper.SetDefault("influxdb.agent_name", "unknown")
}

func InitConfig() {
	var err error
	// config
	addr = fmt.Sprintf("http://%s:%d",
		viper.GetString("influxdb.addr"),
		viper.GetInt("influxdb.port"))

	username = viper.GetString("influxdb.username")
	password = viper.GetString("influxdb.password")
	database = viper.GetString("influxdb.database")
	batchSize = viper.GetInt("influxdb.batch_size")
	batchId = 0

	agentName = viper.GetString("influxdb.agent_name")

	client, err = idb.NewHTTPClient(idb.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password})

	if err != nil {
		log.Fatal().Msg(err.Error())
	} else {
		log.Info().Msg("Connecting to InfluxDB")
	}

	// create database if it does not exist yet
	new_db_query := idb.Query{
		Command:  fmt.Sprintf("CREATE DATABASE %s", database),
		Database: database}
	_, err = client.Query(new_db_query)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// create the batch of records
	batch, err = idb.NewBatchPoints(idb.BatchPointsConfig{
		Database:  database,
		Precision: "us"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func untypeMap(m map[string]float64) map[string]interface{} {
	M := make(map[string]interface{})
	for key, value := range m {
		M[key] = value
	}
	return M
}

func PushRecord(statValues map[string]float64, seriesName string, t time.Time) error {
	point, err := idb.NewPoint(seriesName,
		map[string]string{"agent": agentName},
		untypeMap(statValues), t)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	batch.AddPoint(point)
	batchId++
	if batchId%batchSize == 0 {
		return Flush()
	}
	return nil
}

func Flush() error {
	err := client.Write(batch)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return err
}

// GetAddress returns the network address of the influxdb instance
// ip:port
func GetAddress() string {
	return addr
}

func Close() error {
	Flush()
	log.Info().Msg("Closing InfluxDB connection")
	return client.Close()
}

func main() {}
