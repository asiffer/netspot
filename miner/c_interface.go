// c_interface.go

package miner

import (
	"C"
	"fmt"
)
import "time"

func join(joiner string, list []string) string {
	var joinedStrings string
	l := len(list) - 1
	for i := 0; i < l; i++ {
		joinedStrings += list[i] + joiner
	}
	return joinedStrings + list[l]
}

// CGetAvailableDevices returns the raw list of available devices (raw string)
//export CGetAvailableDevices
func CGetAvailableDevices() *C.char {
	return C.CString(join(",", GetAvailableDevices()))
}

// CGetDevice returns the current device
//export CGetDevice
func CGetDevice() *C.char {
	return C.CString(device)
}

// CSetDevice changes the current device
//export CSetDevice
func CSetDevice(devPtr *C.char) int {
	dev := C.GoString(devPtr)
	if contains(AvailableDevices, dev) {
		device = dev
		iface = true
	} else if fileExists(dev) {
		device = dev
		iface = false
	} else {
		fmt.Println("Unknown device")
		return 1
	}
	return 0
}

// CSetTimeout changes the current timeout (given in seconds)
//export CSetTimeout
func CSetTimeout(t int64) {
	var sec int64 = 1000000000 // 1sec = 1 000 000 000 ns
	timeout = time.Duration(t * sec)
}

// CGetLoadedCounters returns the raw list of loaded counters
//export CGetLoadedCounters
func CGetLoadedCounters() *C.char {
	dl := GetLoadedCounters()
	return C.CString(join(",", dl))
}

// CLoadFromName loads a counter from the given name
//export CLoadFromName
func CLoadFromName(ctrnamePtr *C.char) int {
	ctr := counterFromName(C.GoString(ctrnamePtr))
	id, _ := load(ctr)
	return id
}

// CUnloadFromName unloads a counter from the given name
//export CUnloadFromName
func CUnloadFromName(ctrnamePtr *C.char) int {
	id := idFromName(C.GoString(ctrnamePtr))
	if id == -1 {
		return -1
	}
	Unload(id)
	return 0
}
