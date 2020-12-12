package miner

import (
	"fmt"
	"netspot/miner/counters"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func init() {
}

// TESTING STRUCTURE ======================================= //
// ========================================================= //
// ========================================================= //

// WaitGroup
// type WaitGroup struct {
// 	wg    *sync.WaitGroup
// 	count int64
// }

// func NewWaitGroup() *WaitGroup {
// 	var wg sync.WaitGroup
// 	return &WaitGroup{
// 		wg:    &wg,
// 		count: 0,
// 	}
// }

// func (w *WaitGroup) Add(delta int) {
// 	atomic.AddInt64(&w.count, int64(delta))
// 	w.wg.Add(delta)
// }

// func (w *WaitGroup) Done() {
// 	atomic.AddInt64(&w.count, -1)
// 	w.wg.Done()
// }

// func (w *WaitGroup) Wait() {
// 	w.wg.Wait()
// }

// func (w *WaitGroup) WaitDebug() {
// 	com := make(chan int)
// 	go func() {
// 		w.Wait()
// 		com <- 0
// 	}()

// 	for {
// 		select {
// 		case <-com:
// 			return
// 		default:
// 			fmt.Println(atomic.LoadInt64(&w.count))
// 			time.Sleep(1000 * time.Millisecond)
// 		}
// 	}
// }

// ========================================================= //

// CounterList is a structure which precises
// the types of the counters. It aims to accelerate
// the packet parsing and... it seems to work :)
type CounterList struct {
	pkt   []counters.PktCtrInterface
	ip4   []counters.IPv4CtrInterface
	tcp   []counters.TCPCtrInterface
	udp   []counters.UDPCtrInterface
	icmp4 []counters.ICMPv4CtrInterface
	arp   []counters.ARPCtrInterface
}

// Dispatcher is the main structures which manage
// the counters
type Dispatcher struct {
	pool            sync.WaitGroup
	list            *CounterList
	counters        map[string]counters.BaseCtrInterface
	receivedPackets uint64
}

// NewDispatcher init a new Dispatcher
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		list:     nil,
		counters: make(map[string]counters.BaseCtrInterface),
	}
}

// init must be called at runtime
func (d *Dispatcher) init() {
	d.buildCounterList()
}

// buildCounterList builds the internal
// CounterList from the loaded counters
func (d *Dispatcher) buildCounterList() {
	list := CounterList{
		pkt:   make([]counters.PktCtrInterface, 0),
		ip4:   make([]counters.IPv4CtrInterface, 0),
		tcp:   make([]counters.TCPCtrInterface, 0),
		udp:   make([]counters.UDPCtrInterface, 0),
		icmp4: make([]counters.ICMPv4CtrInterface, 0),
		arp:   make([]counters.ARPCtrInterface, 0),
	}

	for _, ctr := range d.counters {
		switch z := ctr.(type) {
		// NEW
		case counters.ARPCtrInterface:
			list.arp = append(list.arp, z)
		case counters.IPv4CtrInterface:
			list.ip4 = append(list.ip4, z)
		case counters.TCPCtrInterface:
			list.tcp = append(list.tcp, z)
		case counters.UDPCtrInterface:
			list.udp = append(list.udp, z)
		case counters.ICMPv4CtrInterface:
			list.icmp4 = append(list.icmp4, z)
		case counters.PktCtrInterface:
			list.pkt = append(list.pkt, z)
		}
	}

	d.list = &list
}

// load adds a counter to the dispatcher
func (d *Dispatcher) load(name string) error {
	ctr, exists := counters.AvailableCounters[name]
	if !exists {
		return fmt.Errorf("The counter %s does not exists", name)
	}
	d.counters[name] = ctr
	return nil
}

// unload removes a counter
func (d *Dispatcher) unload(name string) error {
	_, exists := counters.AvailableCounters[name]
	if !exists {
		return fmt.Errorf("The counter %s does not exists", name)
	}

	delete(d.counters, name)
	return nil
}

// dissect analyzes the layers of the packets and call
// the right callback
func (d *Dispatcher) dissect(pkt gopacket.Packet) {
	// internal pool
	defer d.pool.Done()

	for _, ctr := range d.list.pkt {
		ctr.Process(pkt)
	}

	ipLayer := pkt.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip4, _ := ipLayer.(*layers.IPv4)
		for _, ctr := range d.list.ip4 {
			ctr.Process(ip4)
		}

		switch t := pkt.Layer(ip4.NextLayerType()).(type) {
		case *layers.TCP:
			for _, ctr := range d.list.tcp {
				ctr.Process(t)
			}

		case *layers.UDP:
			for _, ctr := range d.list.udp {
				ctr.Process(t)
			}

		case *layers.ICMPv4:
			for _, ctr := range d.list.icmp4 {
				ctr.Process(t)
			}
		default:
			//ignore
		}
	} else if arpLayer := pkt.Layer(layers.LayerTypeARP); arpLayer != nil {
		t, _ := arpLayer.(*layers.ARP)
		for _, ctr := range d.list.arp {
			ctr.Process(t)
		}
	}

}

// dispatch
func (d *Dispatcher) dispatch(packet gopacket.Packet) {
	d.pool.Add(1)
	d.receivedPackets++
	go d.dissect(packet)
}

// terminate wait for all the dissect operations
// to finish
func (d *Dispatcher) terminate() {
	d.pool.Wait()
}

// flushAll gets the values of every counter and
// resets them
func (d *Dispatcher) flushAll() map[string]uint64 {
	// flush counters
	data := make(map[string]uint64)
	for name, ctr := range d.counters {
		// get value
		data[name] = ctr.Value()
		// reset counter
		ctr.Reset()
	}
	return data
}

// terminateAndflushAll terminates the goroutines,
// gets the values of every counter and resets them.
func (d *Dispatcher) terminateAndFlushAll() map[string]uint64 {
	// terminate
	d.terminate()
	// flush counters
	return d.flushAll()
}

// getAll gets the values of every counter
func (d *Dispatcher) getAll() map[string]uint64 {
	// flush counters
	data := make(map[string]uint64)
	for name, ctr := range d.counters {
		// get value
		data[name] = ctr.Value()
	}

	return data
}

// loadedCounters returns the list of the loaded counters
func (d *Dispatcher) loadedCounters() []string {
	lc := make([]string, len(d.counters))
	i := 0
	for name := range d.counters {
		lc[i] = name
		i++
	}
	return lc
}
