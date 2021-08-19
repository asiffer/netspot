---
title: Exporting modules
weight: -10
summary: Depending on your data/logs storage, `netspot` can be adapted to your needs.
---

The exporting modules are responsible of sending `netspot` data to the desired backend: console, file, database etc.

They basically receive the computed statistics and the possible alarms raised by the `analyzer` module (see the [architecture](../../get-started/architecture/index.md)). The `netspot/exporter` package manages the exporting modules.

The exporting modules are located in the `exporter` subpackage. One module may merely consist in two files: source and test.

## Description

In the code side, an exporting module is anything that implements the `ExportingModule` interface.

```go
// ExportingModule is the general interface which denotes
// a module which sends data from netspot
type ExportingModule interface {
	// Return the name of the module
	Name() string
	// Init the module (init some variables)
	Init() error
	// Start the module (make it ready to receive data)
	Start(series string) error
	// Aimed to write raw statistics
	Write(time.Time, map[string]float64) error
	// Aimed to write alerts
	Warn(time.Time, *SpotAlert) error
	// Close the module
	Close() error
}
```

Let us detail the functions to implement:

- `Name` returns the name of the module. This name is used in the configuration so it must be clear and simple. One should avoid non-letter
  characters (like space, dash, underscores, numbers...) but this is not mandatory.
- `Init` is called when `netspot` loads the configuration file: it asks the `exporter` to init its modules which then runs their `Init` function. This function aims to read the specific config of the module and is reponsible to load it (it may be strange but this the current behavior).
- `Start` is called before `netspot` starts to read to a packets source. It provides a `string` parameters that denotes the name of the run. For example, it may be the name of the network interface combined with a timestamp or the name of a pcap file. This piece of information aims to identify a run. During this step, the handles to send data should be created (database connection, files etc.).
- `Write` is called when statistics are computed so they can be stored. It consumes a timestamp and a map where keys are the name of the statistics and values are their... computed values. This action happens at the frequency given in input (analyzer section of the config file)
- `Warn` is triggered in case of alarm. When the SPOT algorithm finds an abnormal stat value, it fills a `SpotAlert` structure (see below) that is then forwarded to the `exporter` and its loaded modules.
- `Close` is called when a run is finished. This function is likely to close connection or any handle required to export data.

<!-- prettier-ignore -->
!!! info
    The `SpotAlert` structure is declared in the `exporter` package. The `analyzer` fills it when an abnormal event occurs.

    ```go
    // SpotAlert is a simple structure to log alerts sent
    // by spot instances
    type SpotAlert struct {
        Status      string  // "UP_ALERT" or "DOWN_ALERT"
        Stat        string  // Stat is the name of the statistic
        Value       float64 // Value is the [abnormal] value if the statistic
        Code        int     // Code is a return code of the SPOT algorithms (useless here)
        Probability float64 // Probability corresponds to the probability to see an event at least as extreme (higher or lower depending on the Status)
    }
    ```

    It consist in few fields, the most important of which are: the `Status` which says whether the value if too high or too low, the `Value` that is the actual value of the stat (redundant with the input of `Write`) and `Probability` that is the probability to see an event at least as extreme.

## Example

### Structure and registration

Let us show the **File** exporting module. First we declare a new structure with some fields we will need.

```go
// File is the file logger
type File struct {
	data             bool
	alarm            bool
	dataAddress      string
	alarmAddress     string
	seriesName       string
	dataFileHandler  *os.File
	alarmFileHandler *os.File
}
```

In the `init` function, we must register the module and its required parameters (they will be passed through config).

```go
func init() {
    // register the module
	Register(&File{})
    // register some parameters
	RegisterParameter("file.data", nil, "File to export data")
	RegisterParameter("file.alarm", nil, "File to export the alarms")
}
```

Here we declare the module, that is the object that implements the `ExportingModule` interface. Then we register some parameters through the `RegisterParameter` function. Its prototype is given below.

```go
// RegisterParameter is a helper function to
// make exporting module parameters available.
// The name must have the form "<module name>.<parameter>"
func RegisterParameter(name string, value interface{}, usage string)
```

The `name` corresponds to the key in the config file. For the `File` exporting module, we assume a `file` namespace inside the `exporter` section in the config file. The parameters are then some keys within this namespace. In the TOML case, the configuration file will look like this:

```ini
# netspot.toml
[exporter.file]
data =
alarm =
```

The `value` parameter is the default value if it is not provided. It has no strong type as it naturally depends on what you need (number, time, string ...). You can put eveything you want for this parameter insofar as you will handle it (in the `Init` function especially).

Finally the `usage` parameter should provide a description of the parameter.

### Initialization

For the **File** exporting module, the initialization looks for the keys in the configuration. First we check if the keys exist (with `HasNotNilKey`). If `exporter.file.data` is missing, the exporting module will do nothing when `Write` is called. The `exporter.file.alarm` key is similar but for the `Warn` function.
The module is loaded if at least one of both keys is set.

```go
// Init reads the config of the modules
func (f *File) Init() error {
	var err error

	f.data = config.HasNotNilKey("exporter.file.data")
	f.alarm = config.HasNotNilKey("exporter.file.alarm")

	if f.data {
		f.dataAddress, err = config.GetPath("exporter.file.data")
		if err != nil {
			return err
		}
	}

	if f.alarm {
		f.alarmAddress, err = config.GetPath("exporter.file.alarm")
		if err != nil {
			return err
		}
	}

	if f.data || f.alarm {
		return Load(f.Name())
	}
	return nil
}
```

### Start/Close

When `netspot` starts the analysis, it asks the `exporter` to starts its modules. For the **File** module, files are open according to the configuration. If every module starts well (meaning that the backend is OK), the analysis can start (errors are sent otherwise).

```go
// Start generate the connection from the module to the endpoint
func (f *File) Start(series string) error {
	var err error
	f.seriesName = series
	f.updateFileFromSeriesName()

	// init file handlers
	// data logger
	if f.data {
		if f.dataFileHandler, err = os.Create(f.dataAddress); err != nil {
			return err
		}
	}
	// alarm logger
	if f.alarm {
		if f.alarmFileHandler, err = os.Create(f.alarmAddress); err != nil {
			return err
		}
	}
	return nil
}
```

In the end of the analysis, after every packet has been processed, `netspot` asks the exporter to close its modules. In our case, it merely closes the file handlers that have been previously created.

```go
// Close the file handles
func (f *File) Close() error {
	if f.alarmFileHandler != nil {
		if err := f.alarmFileHandler.Close(); err != nil {
			return fmt.Errorf("error while closing '%s' module (%v)", f.Name(), err)
		}
	}
	if f.dataFileHandler != nil {
		if err := f.dataFileHandler.Close(); err != nil {
			return fmt.Errorf("error while closing '%s' module (%v)", f.Name(), err)
		}
	}
	return nil
}
```

### Data/Alarm handling

The `Write` and `Warn` functions are generally simple, and you may only need to properly format the data before sending it. The `exporter` packages includes some helpers like `jsonifyWithTime(time.Time, map[string]float64) string` that turns the input into a simple JSON record. For the `SpotAlert` structure, there is the analogous `toJSONwithTime(time.Time)` method.

```go
// Write logs data
func (f *File) Write(t time.Time, data map[string]float64) error {

	if f.data {
		f.dataFileHandler.WriteString(jsonifyWithTime(t, data))
		f.dataFileHandler.Write([]byte{'\n'})
	}
	return nil
}

// Warn logs alarms
func (f *File) Warn(t time.Time, s *SpotAlert) error {
	if f.alarm {
		f.alarmFileHandler.WriteString(s.toJSONwithTime(t))
		f.alarmFileHandler.Write([]byte{'\n'})
	}
	return nil
}
```

<!-- prettier-ignore -->
!!! warning
    In these two functions, errors can be better handled since the results of `WriteString` and `Write` are not checked.
