// influxdb1.go

package exporter

import (
	"fmt"
	"netspot/config"
	"regexp"
	"time"

	influx "github.com/influxdata/influxdb1-client/v2"
)

const (
	defaultBatchSize = 10
)

// InfluxDB is sends data to an InfluxDB database (v1)
type InfluxDB struct {
	data       bool
	alarm      bool
	address    string
	database   string
	username   string
	password   string
	seriesName string
	client     influx.Client
	batch      influx.BatchPoints
	agentName  string
	batchSize  int
	batchID    int
}

func init() {
	RegisterAndSetDefaults(&InfluxDB{}, map[string]interface{}{
		"exporter.influxdb.address":    "http://127.0.0.1:8086",
		"exporter.influxdb.database":   "netspot",
		"exporter.influxdb.username":   "netspot",
		"exporter.influxdb.password":   "netspot",
		"exporter.influxdb.batch_size": 10,
		"exporter.influxdb.agent_name": "local",
	})
}

// Main functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

// Name returns the name of the exporter
func (i *InfluxDB) Name() string {
	return "influxdb"
}

// Options return the parameters of the shipper
// func (i *InfluxDB) Options() map[string]string {
// 	return map[string]string{
// 		"data":  "A boolean to activate console data logging",
// 		"alarm": "A boolean to activate console alarm logging",
// 		"address": `A string which gives the full address of the influxdb instance.
// Ex: http://localhost:8086`,
// 		"database":   "A string which sets the database to use",
// 		"username":   "A string which defines the user which connects to the database",
// 		"password":   "A string which defines the corresponding password",
// 		"batch_size": "A positive integer which sets the number of records which are sent in a row",
// 		"agent_name": "A string which gives an additional tag to the records",
// 	}
// }

// Status return the status of the shipper
func (i *InfluxDB) Status() map[string]interface{} {
	m := map[string]interface{}{
		"batch_size": i.batchSize,
		"agent_name": i.agentName,
	}
	if i.data || i.alarm {
		m["address"] = i.address
		m["username"] = i.username
		m["database"] = i.database
	}
	return m
}

// Init reads the config of the module
func (i *InfluxDB) Init() error {
	var err error
	// TODO init with data and alarm (->different tags)
	i.data = config.MustBool("exporter.influxdb.data")
	i.alarm = config.MustBool("exporter.influxdb.alarm")

	// agent
	i.agentName, err = config.GetString("exporter.influxdb.agent_name")
	if err != nil {
		return err
	}

	// batch
	i.batchSize, err = config.GetInt("exporter.influxdb.batch_size")
	if err != nil {
		return err
	}

	// set ID to zero
	i.batchID = 0

	// address, user, password, database
	i.address, err = config.GetString("exporter.influxdb.address")
	if err != nil {
		return err
	}

	i.username, err = config.GetString("exporter.influxdb.username")
	if err != nil {
		return err
	}

	i.password, err = config.GetString("exporter.influxdb.password")
	if err != nil {
		return err
	}

	i.database, err = config.GetString("exporter.influxdb.database")
	if err != nil {
		return err
	}

	if i.data || i.alarm {
		return Load(i.Name())
	}

	return nil
}

// Start generate the connection from the shipper to the endpoint
func (i *InfluxDB) Start(series string) error {
	var err error
	i.seriesName = series

	// client
	i.client, err = influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     i.address,
		Username: i.username,
		Password: i.password,
	})
	if err != nil {
		return fmt.Errorf("Error while initializing InfluxDB client (%v)", err)
	}

	// create database (a valid client is required)
	err = i.checkDatabase()
	if err != nil {
		return fmt.Errorf("Error while creating InfluxDB database '%s' (%v)", i.database, err)
	}

	// create the batch of records
	i.batch, err = influx.NewBatchPoints(influx.BatchPointsConfig{
		Database: i.database})
	if err != nil {
		return fmt.Errorf("Error while creating InfluxDB batch of data points (%v)", err)
	}
	return nil
}

// Write logs data
func (i *InfluxDB) Write(t time.Time, data map[string]float64) error {
	if i.data {
		point, err := influx.NewPoint(
			i.seriesName,
			map[string]string{"agent": i.agentName, "type": "data"},
			untypeMap(data),
			t)
		if err != nil {
			return err
		}
		i.batch.AddPoint(point)
		i.batchID++
		if i.batchID%i.batchSize == 0 {
			return i.flush()
		}
	}
	return nil
}

// Warn logs alarms
func (i *InfluxDB) Warn(t time.Time, s *SpotAlert) error {
	if i.alarm {
		point, err := influx.NewPoint(
			i.seriesName,
			map[string]string{
				"agent": i.agentName,
				"type":  "alarm",
			},
			map[string]interface{}{
				"status":      s.Status,
				"value":       s.Value,
				"code":        s.Code,
				"probability": s.Probability,
			},
			t)
		if err != nil {
			return err
		}
		i.batch.AddPoint(point)
		i.batchID++
		if i.batchID%i.batchSize == 0 {
			return i.flush()
		}
	}
	return nil
}

// Close flush the batch of points and close client
func (i *InfluxDB) Close() error {
	if i.client != nil {
		if err := i.flush(); err != nil {
			return err
		}
		if err := i.client.Close(); err != nil {
			return err
		}
	}
	return nil
}

// LogsData tells whether the shipper logs data
func (i *InfluxDB) LogsData() bool {
	return i.data
}

// LogsAlarm tells whether the shipper logs alarm
func (i *InfluxDB) LogsAlarm() bool {
	return i.alarm
}

// Side functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

func (i *InfluxDB) checkDatabase() error {
	// avoid injection
	database, err := sanitizeDB(i.database)
	if err != nil {
		return fmt.Errorf("Error while sanitizing database name (%v)", err)
	}

	// create database if it does not exist yet
	newDBQuery := influx.Query{
		Command:  fmt.Sprintf("CREATE DATABASE %s", database),
		Database: i.database}
	_, err = i.client.Query(newDBQuery)
	return err
}

// Flush sends the batch of points to influxdb1.
func (i *InfluxDB) flush() error {
	if err := i.client.Write(i.batch); err != nil {
		return fmt.Errorf("Error while writing batch of data (%v)", err)
	}
	// reset
	i.batchID = 0
	return nil
}

func sanitizeDB(db string) (string, error) {
	// regex to say we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(db, ""), nil
}
