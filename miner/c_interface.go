// c_interface.go
package miner

import (
	"C"
	"fmt"
)

func join(joiner string, list []string) string {
	var joinedStrings string = ""
	l := len(list) - 1
	for i := 0; i < l; i++ {
		joinedStrings += list[i] + joiner
	}
	return joinedStrings + list[l]
}

//export C_GetAvailableDevices
func C_GetAvailableDevices() *C.char {
	return C.CString(join(",", GetAvailableDevices()))
}

//export C_GetDevice
func C_GetDevice() *C.char {
	return C.CString(device)
}

//export C_SetDevice
func C_SetDevice(dev_ptr *C.char) int {
	dev := C.GoString(dev_ptr)
	if AvailableDevices.contains(dev) {
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

//export C_GetLoadedCounters
func C_GetLoadedCounters() *C.char {
	dl := GetLoadedCounters()
	return C.CString(join(",", dl))
}

//export C_LoadFromName
func C_LoadFromName(ctrname_ptr *C.char) int {
	ctr := counterFromName(C.GoString(ctrname_ptr))
	id, _ := load(ctr)
	return id
}

//export C_UnloadFromName
func C_UnloadFromName(ctrname_ptr *C.char) int {
	id := idFromName(C.GoString(ctrname_ptr))
	if id == -1 {
		return -1
	} else {
		Unload(id)
		return 0
	}
}
