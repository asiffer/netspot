// netspotctl.go

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type NullArgs struct{}

var (
	client          *rpc.Client
	interfaces      []string
	connectionError error
	connected       string = "200 Connected to Go RPC" // (rpc package) Can connect to RPC service using HTTP CONNECT to rpcPath.
)

//------------------------------------------------------------------------------
// PRINTING FUNCTIONS
//------------------------------------------------------------------------------

func printOK(s string) {
	fmt.Printf("\033[32m%s\033[0m\n", s)
}

func printfOK(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("\033[32m%s\033[0m\n", msg)
}

func printWARNING(s string) {
	fmt.Printf("\033[1;33m%s\033[0m\n", s)
}

func printfWARNING(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("\033[1;33m%s\033[0m\n", msg)
}

func printERROR(s string) {
	fmt.Printf("\033[31m%s\033[0m\n", s)
}

func printfERROR(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("\033[31m%s\033[0m\n", msg)
}

//------------------------------------------------------------------------------
// INITIALIZER
//------------------------------------------------------------------------------

// InitConsoleWriter initializes the console outputing details about the
// netspotctl events.
func InitConsoleWriter() {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMicro}

	output.FormatLevel = func(i interface{}) string {
		switch fmt.Sprintf("%s", i) {
		case "warn":
			return "\033[31mWARNING\033[0m"
		case "info":
			return "\033[32m   INFO\033[0m"
		case "fatal":
			return "\033[1m\033[31m  FATAL\033[0m"
		case "error":
			return "\033[1m\033[31m  ERROR\033[0m"
		case "debug":
			return "\033[33m  DEBUG\033[0m"
		case "panic":
			return "\033[1m\033[31m  PANIC\033[0m"
		default:
			return fmt.Sprintf("%s", i)
		}
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		field := fmt.Sprintf("%s", i)
		switch field {
		case "type", "module":
			return ""
		default:
			return "\033[2m\033[37m" + field + ":" + "\033[0m"
		}
	}

	output.FormatFieldValue = func(i interface{}) string {
		switch i.(type) {
		case float64:
			f := i.(float64)
			if f < 1e-3 {
				return fmt.Sprintf("%e", f)
			} else {
				return fmt.Sprintf("%.5f", f)
			}
		case int32, int16, int8, int:
			return fmt.Sprintf("%d", i)
		default:
			return strings.ToUpper(fmt.Sprintf("%s", i))
		}
	}

	output.PartsOrder = []string{"time", "level", "message"}
	log.Logger = log.Output(output)
	zerolog.TimeFieldFormat = time.StampNano
}

func Init() {
	// var err error
	viper.SetConfigName("netspotctl")
	viper.AddConfigPath("/etc/netspotctl/")
	viper.AddConfigPath("/home/asr/Documents/Work/go/src/netspot/netspotctl")
	viper.AddConfigPath(".")

	viper.SetDefault("netspot.host", "localhost")
	viper.SetDefault("netspot.port", 11000)

	viper.ReadInConfig()
	host := viper.GetString("netspot.host")
	port := viper.GetInt("netspot.port")
	addr := fmt.Sprintf("%s:%d", host, port)
	// client, err = rpc.DialHTTP("tcp", addr)
	// if err != nil {

	// 	printfERROR("Error: %s", err.Error())
	// 	// os.Exit(1)
	// } else {
	// 	// fmt.Printf("Connected to netspot server (%s)\n", addr)
	// 	printfOK("Connected to netspot server (%s)", addr)
	// }
	execConnect(addr)
}

//------------------------------------------------------------------------------
// SIDE FUNCTIONS
//------------------------------------------------------------------------------

func find(a []string, s string) int {
	for i, v := range a {
		if v == s {
			return i
		}
	}
	return -1
}

// sub returns a - b
func sub(a []string, b []string) []string {
	out := make([]string, 0)
	for _, v := range a {
		if find(b, v) == -1 {
			out = append(out, v)
		}
	}
	return out
}

func basicInput(msg string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(msg)
	text, _ := reader.ReadString('\n')
	return strings.Replace(text, "\n", "", -1)
}

func availableStats() []string {
	var i int
	stats := make([]string, 0)
	// err := client.Call("Netspot.ListAvailable", &i, &stats)
	err := call("Netspot.ListAvailable", &i, &stats)
	if err != nil {
		printERROR(err.Error())
	}
	return stats
}

func loadedStats() []string {
	var i int
	stats := make([]string, 0)
	// err := client.Call("Netspot.ListLoaded", &i, &stats)
	err := call("Netspot.ListLoaded", &i, &stats)
	if err != nil {
		printERROR(err.Error())
	}
	return stats
}

func availableInterfaces() []string {
	var i int
	ifaces := make([]string, 0)
	// err := client.Call("Netspot.AvailableInterface", &i, &ifaces)
	err := call("Netspot.AvailableInterface", &i, &ifaces)
	if err != nil {
		printERROR(err.Error())
	}
	return ifaces
}

// The two folowing functions are take from net/rpc package.
// The DialHTTP function does not use timeout natively while
// net package includes a dedicated function. So these two function
// add timeout functionnality.

// Dial with Timeout
func DialHTTPTimeout(network, address string, timeout time.Duration) (*rpc.Client, error) {
	return DialHTTPPathTimeout(network, address, rpc.DefaultRPCPath, timeout)
}

// DialHTTPPath connects to an HTTP RPC server
// at the specified network address and path.
func DialHTTPPathTimeout(network, address, path string, timeout time.Duration) (*rpc.Client, error) {
	var err error
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return rpc.NewClient(conn), nil
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	conn.Close()

	return nil, &net.OpError{
		Op:   "dial-http",
		Net:  network + " " + address,
		Addr: nil,
		Err:  err,
	}

}

// hook rpc.Client.Call function
func call(serviceMethod string, args interface{}, reply interface{}) error {
	if connectionError == nil {
		return client.Call(serviceMethod, args, reply)
	} else {
		return errors.New("You are not connected to a NetSpot server")
	}
}

//------------------------------------------------------------------------------
// ACTION FUNCTIONS
//------------------------------------------------------------------------------

func execConnect(addr string) {
	var err error
	// client, err = rpc.DialHTTP("tcp", addr)
	timeout := 5 * time.Second
	client, err = DialHTTPTimeout("tcp", addr, timeout)
	if err != nil {
		printfERROR("Error: %s", err.Error())
	} else {
		printfOK("Connected to netspot server (%s)", addr)
	}
	connectionError = err
}

func execParseConnect(addr []string) {
	switch len(addr) {
	case 1:
		execConnect(addr[0])
	case 2:
		a := fmt.Sprint(addr[0], ":", addr[1])
		execConnect(a)
	default:
		printERROR("Usage: connect [ip]:[port] (or [ip] [port])")
	}
}

func execExit() {
	answer := basicInput("Leave? (Y/n) ")
	switch answer {
	case "n", "N", "no", "No":
		return
	default:
		os.Exit(0)
	}
}

func execConfig() {
	var s string
	var i int
	// err := client.Call("Netspot.Config", &i, &s)
	err := call("Netspot.Config", &i, &s)
	if err != nil {
		printERROR(err.Error())
	} else {
		fmt.Println(s)
	}
}

func execStart() {
	var none, i int
	// err := client.Call("Netspot.Start", &none, &i)
	err := call("Netspot.Start", &none, &i)
	if err == nil {
		printOK("Start sniffing...")
	} else {
		printERROR(err.Error())
	}
}

func execStop() {
	var none, i int
	// err := client.Call("Netspot.Stop", &none, &i)
	err := call("Netspot.Stop", &none, &i)
	if err == nil {
		printOK("Stop sniffing...")
	} else {
		printERROR(err.Error())
	}
}

func execLoad(s string) {
	var i int
	// err := client.Call("Netspot.Load", s, &i)
	err := call("Netspot.Load", s, &i)
	if err == nil {
		printfOK("%s statistics loaded", s)
	} else {
		printERROR(err.Error())
	}
}

func execUnload(s string) {
	var i int
	var err error
	if s == "all" || s == "All" {
		err = call("Netspot.UnloadAll", s, &i)
		if err == nil {
			printOK("All statistics unloaded")
		} else {
			printERROR(err.Error())
		}
	} else {
		// err := client.Call("Netspot.Unload", s, &i)
		err = call("Netspot.Unload", s, &i)
		if err == nil {
			printfOK("%s statistics unloaded", s)
		} else {
			printERROR(err.Error())
		}
	}
}

func execList() {
	ls := loadedStats()
	as := availableStats()
	for _, r := range as {
		if find(ls, r) == -1 {
			fmt.Println(r)
		} else {
			fmt.Printf("\033[1;34m%s\033[0m\n", r)
		}
	}
}

func execReset() {
	var i, j int
	// err := client.Call("Netspot.Zero", &i, &j)
	err := call("Netspot.Zero", &i, &j)
	if err == nil {
		printfOK("Netspot has been reset")
	} else {
		printERROR(err.Error())
	}
}

func execShow(s string) {
	var status string
	// err := client.Call("Netspot.StatStatus", s, &status)
	err := call("Netspot.StatStatus", s, &status)
	if err == nil {
		fmt.Println(status)
	} else {
		printERROR(err.Error())
	}
}

func execSetDevice(dev string) {
	var i int
	// err := client.Call("Netspot.SetDevice", dev, &i)
	err := call("Netspot.SetDevice", dev, &i)
	if err == nil {
		printfOK(`Set device to "%s"`, dev)
	} else {
		printERROR(err.Error())
	}
}

func execSetPromiscuous(b bool) {
	var i int
	// err := client.Call("Netspot.SetPromiscuous", b, &i)
	err := call("Netspot.SetPromiscuous", b, &i)
	if err == nil {
		printfOK("Set promiscuous to %v", b)
	} else {
		printERROR(err.Error())
	}
}

func execSetPeriod(duration string) {
	var i int
	// err := client.Call("Netspot.SetPeriod", duration, &i)
	err := call("Netspot.SetPeriod", duration, &i)
	if err == nil {
		d, _ := time.ParseDuration(duration)
		printfOK("Period set to %s", d)
	} else {
		printERROR(err.Error())
	}
}

func execIface() {
	interfaces = availableInterfaces()
	for _, iface := range interfaces {
		fmt.Println(iface)
	}
}

func executor(s string) {
	words := strings.Split(s, " ")
	// First word
	switch words[0] {
	case "exit":
		execExit()
	case "reset":
		execReset()
	case "config":
		execConfig()
	case "iface":
		execIface()
	case "start":
		execStart()
	case "stop":
		execStop()
	case "connect":
		if len(words) > 1 {
			execParseConnect(words[1:])
		} else {
			printERROR("Usage: connect [ip]:[port] (or [ip] [port])")
		}
	case "load":
		if len(words) > 1 {
			execLoad(words[1])
		} else {
			printERROR("Usage: show [statistics]")
		}
	case "unload":
		if len(words) > 1 {
			execUnload(words[1])
		} else {
			printERROR("Usage: unload [statistics]")
		}
	case "list":
		execList()
	case "show":
		if len(words) > 1 {
			execShow(words[1])
		} else {
			printERROR("Missing statistics")
		}
	case "set":
		if len(words) > 2 {
			switch words[1] {
			case "device":
				execSetDevice(words[2])
			case "promisc":
				switch words[2] {
				case "yes", "Yes", "true", "True", "y", "Y":
					execSetPromiscuous(true)
				case "no", "No", "false", "False", "n", "N":
					execSetPromiscuous(false)
				default:
					printERROR("Usage: set promisc [true/false]")
				}
			case "period":
				execSetPeriod(words[2])
			default:
				printERROR("Option not available")
			}
		} else {
			printERROR("Usage: set [option] [value]")
		}
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	var s []prompt.Suggest
	words := strings.Split(d.Text, " ")

	switch len(words) {
	case 1:
		s = []prompt.Suggest{
			{Text: "connect", Description: "Connect to a NetSpot server"},
			{Text: "config", Description: "netspot configuration"},
			{Text: "exit", Description: "Leave netspot"},
			{Text: "iface", Description: "List all sniffable interfaces"},
			{Text: "list", Description: "List all statistics"},
			{Text: "load", Description: "Load a new statistics"},
			{Text: "reset", Description: "Reset Netspot"},
			{Text: "show", Description: "Details about a loaded statistics"},
			{Text: "start", Description: "Start monitoring"},
			{Text: "stop", Description: "Stop monitoring"},
			{Text: "unload", Description: "Unload a statistics"},
		}
	case 2:
		switch words[0] {
		case "show":
			for _, r := range loadedStats() {
				s = append(s, prompt.Suggest{Text: r})
			}

		case "load":
			remain := sub(availableStats(), loadedStats())
			for _, r := range remain {
				s = append(s, prompt.Suggest{Text: r})
			}
		case "unload":
			for _, r := range loadedStats() {
				s = append(s, prompt.Suggest{Text: r})
			}
			s = append(s, prompt.Suggest{Text: "all", Description: "Unload all statistics"})
		case "set":
			s = []prompt.Suggest{
				{Text: "device", Description: "Set the device to listen (file of pcap)"},
				{Text: "promisc", Description: "Set the promiscuous mode"},
				{Text: "period", Description: "Set the period for stat compuation"}}
		case "iface":
			for _, r := range interfaces {
				s = append(s, prompt.Suggest{Text: r})
			}
		}
	case 3:
		switch words[0] {
		case "set":
			switch words[1] {
			case "promisc":
				s = []prompt.Suggest{
					{Text: "true", Description: "Activate promiscuous mode"},
					{Text: "false", Description: "Desctivate promiscuous mode"}}
			case "device":
				for _, r := range interfaces {
					s = append(s, prompt.Suggest{Text: r})
				}
			}
		}
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {
	InitConsoleWriter()
	Init()

	if connectionError == nil {
		interfaces = availableInterfaces()
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("> "),
		prompt.OptionPrefixTextColor(prompt.White))

	p.Run()
}
