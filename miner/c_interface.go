// c_interface.go

package miner

/*
#include <stdint.h>
extern char * CGetAvailableDevices();
extern char * CGetDevice();
extern int CSetDevice(char* s);
extern void CSetTimeout(int i);
extern uint64_t CGetCounterValue(int id);
*/
import "C"

import (
	"time"

	"github.com/rs/zerolog/log"
)

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
func CSetDevice(devPtr *C.char) C.int {
	dev := C.GoString(devPtr)
	if contains(AvailableDevices, dev) {
		device = dev
		iface = true
	} else if fileExists(dev) {
		device = dev
		iface = false
	} else {
		//fmt.Println("Unknown device")
		return C.int(1)
	}
	return C.int(0)
}

// CSetTimeout changes the current timeout (given in seconds)
//export CSetTimeout
func CSetTimeout(t C.int) {
	timeout = time.Duration(int64(t) * int64(time.Second))
}

// CGetLoadedCounters returns the raw list of loaded counters
//export CGetLoadedCounters
func CGetLoadedCounters() *C.char {
	dl := GetLoadedCounters()
	return C.CString(join(",", dl))
}

// CLoadFromName loads a counter from the given name
//export CLoadFromName
func CLoadFromName(ctrnamePtr *C.char) C.int {
	ctr := counterFromName(C.GoString(ctrnamePtr))
	id, _ := load(ctr)
	return C.int(id)
}

// CUnloadFromName unloads a counter from the given name
//export CUnloadFromName
func CUnloadFromName(ctrnamePtr *C.char) C.int {
	id := idFromName(C.GoString(ctrnamePtr))
	if id == -1 {
		return -1
	}
	Unload(id)
	return 0
}

// CGetCounterValue returns the current value of the counter
// identified by its id
//export CGetCounterValue
func CGetCounterValue(cid C.int) C.uint64_t {
	// mux.Lock()
	id := int(cid)
	ctr, ok := counterMap[id]
	if !ok {
		log.Fatal().Msg("Invalid counter identifier")
	}
	if ctr.IsRunning() {
		// send the signal
		counterMap[id].SigPipe() <- uint8(1)
		// return the value
		// defer mux.Unlock()
		return C.uint64_t(<-counterMap[id].ValPipe())
	}
	// defer mux.Unlock()
	return C.uint64_t(counterMap[id].Value())

}

// TESTS

func testCGetAvailableDevices() bool {
	if C.GoString(C.CGetAvailableDevices()) == join(",", AvailableDevices) {
		return true
	}
	return false
}

func testCGetDevice() bool {
	if C.GoString(C.CGetDevice()) != AvailableDevices[0] {
		return true
	}
	return false
}

func testCSetDevice() bool {
	i := CSetDevice(C.CString(AvailableDevices[1]))
	j := CSetDevice(C.CString("_WTF_"))
	if i != 0 || j == 0 {
		return false
	}
	if C.GoString(CGetDevice()) == AvailableDevices[1] {
		return true
	}
	return false
}

func testCSetTimeout() bool {
	CSetTimeout(23)
	if timeout == 23*time.Second {
		return true
	}
	return false
}

func testCGetLoadedCounters() bool {
	UnloadAll()
	LoadFromName("ACK")
	LoadFromName("ICMP")
	s := C.GoString(CGetLoadedCounters())
	if s == join(",", GetLoadedCounters()) {
		return true
	}
	return false
}

func testCLoadFromName() bool {
	id := CLoadFromName(C.CString("SYN"))
	if id >= 0 && contains(GetLoadedCounters(), "SYN") {
		return true
	}
	return false
}

func testCUnloadFromName() bool {
	i := CUnloadFromName(C.CString("SYN"))
	j := CUnloadFromName(C.CString("_WTF_"))
	if i == 0 && j == -1 {
		return true
	}
	return false
}

func testCGetCounterValue(id int, val uint64) bool {
	i := C.int(id)
	v := uint64(CGetCounterValue(i))
	k, _ := GetCounterValue(id)
	if v == k && v == val {
		return true
	}
	return false
}
