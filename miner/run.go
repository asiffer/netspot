// run.go

package miner

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"netspot/miner/counters"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

//----------------------------------------------------------------------------//
//---------------------------- EXPORTED FUNCTIONS ----------------------------//
//----------------------------------------------------------------------------//

// IsSniffing returns the sniffing status
func IsSniffing() bool {
	return sniffing
}

// GetSourceTime returns the time given by the current packet source
func GetSourceTime() int64 {
	return SourceTime.UnixNano()
}

// StartSniffing starts to sniff the current device. It does nothing
// if the sniffing is in progress. This is a goroutine, so it returns
// once the sniffing has started.
func StartSniffing() (EventChannel, DataChannel) {
	if !sniffing {
		minerLogger.Info().Msgf("Start sniffing %s", device)
		minerLogger.Debug().Msg(fmt.Sprint("Loaded counters: ", GetLoadedCounters()))
		ec, dc := GoSniff()
		for !sniffing {
			// wait for sniffing
		}
		// update the default channels
		defaultEventChannel = ec
		defaultDataChannel = dc
		return ec, dc
	}
	minerLogger.Debug().Msg("StartSniffing requested but already sniffing")
	return defaultEventChannel, defaultDataChannel
}

// StartSniffingAndWait sniff the current device but does not
// return immediately. This is more relevant when the device
// is a capture file: the sniff stops when all the packets have
// been read.
func StartSniffingAndWait() {
	if !sniffing {
		minerLogger.Info().Msgf("Start sniffing %s", device)
		minerLogger.Debug().Msg(fmt.Sprint("Loaded counters: ", GetLoadedCounters()))
		// events = make(chan uint8)
		Sniff()
	}
	minerLogger.Info().Msg("Sniffing stopped")
}

// SniffAndSendUnix start sniffing and send periodically
// counter values to a unix socket
func SniffAndSendUnix(d time.Duration, reset bool, socket string) {
	// connect to the UNIX socket
	conn, err := net.Dial("unixgram", socket)
	if err != nil {
		minerLogger.Fatal().Msg(err.Error())
	}
	defer conn.Close()
	// initialize the data encoder. It will directly
	// send the data to the UNIX socket
	encoder := gob.NewEncoder(conn)
	// start the sniffing process (get the channels)
	_, dc := GoSniffAndYieldChannel(d)

	// loop
	for {
		select {
		// wait for a tick event to send a snapshot
		case m := <-dc:
			// check whether data have been sent
			// nil means that the connection will be closed
			if m != nil {
				minerLogger.Debug().Msgf("%v", m)
				// encode and send data
				encoder.Encode(m)
			} else {
				// close the connection
				minerLogger.Warn().Msg("Sending EOF")
				conn.Write([]byte{'E', 'O', 'F'})
				return
			}
		}
	}
}

// StopSniffing stops to sniff the device
func StopSniffing() {
	if sniffing {
		defaultEventChannel <- STOP
	}
}

// GetCounterValue returns the current value of the counter
// identified by its id
func GetCounterValue(id int) (uint64, error) {
	ctr, ok := counterMap[id]
	if !ok {
		minerLogger.Error().Msg("Invalid counter identifier")
		return 0, errors.New("Invalid counter identifier")
	}
	if ctr.IsRunning() {
		// send the signal (Get)
		counterMap[id].SigPipe() <- counters.GET
		// return the value
		return <-counterMap[id].ValPipe(), nil
	}
	return counterMap[id].Value(), nil
}

// GetCounterValueAndReset returns the current value of the counter
// identified by its id and reset its value.
func GetCounterValueAndReset(id int) (uint64, error) {
	ctr, ok := counterMap[id]
	if !ok {
		minerLogger.Error().Msg("Invalid counter identifier")
		return 0, errors.New("Invalid counter identifier")
	}
	if ctr.IsRunning() {
		// send the signal (Get+Reset)
		counterMap[id].SigPipe() <- counters.FLUSH
		// return the value
		return <-counterMap[id].ValPipe(), nil
	}
	defer counterMap[id].Reset()
	return counterMap[id].Value(), nil
}

func getValues(pool *sync.WaitGroup) map[int]uint64 {
	m := make(map[int]uint64)
	if pool != nil {
		pool.Wait()
	}
	valmux.Lock()
	for i := range counterMap {
		if v, err := GetCounterValue(i); err == nil {
			m[i] = v
		} else {
			minerLogger.Error().Msg(err.Error())
		}
	}
	valmux.Unlock()
	// st := GetSourceTime()
	// if st > 0 {
	// 	m[0] = uint64(st)
	// } else {
	// 	m[0] = 0
	// }

	// // here we add the information about sent data (counters)
	// m[TypeIndex] = CounterType
	return m
}

func getValuesAndReset(pool *sync.WaitGroup) map[int]uint64 {
	m := make(map[int]uint64)
	// wait until all counters finish
	if pool != nil {
		pool.Wait()
	}
	// Lock the map
	valmux.Lock()
	for i := range counterMap {
		if v, err := GetCounterValueAndReset(i); err == nil {
			m[i] = v
		} else {
			minerLogger.Error().Msg(err.Error())
		}
	}
	valmux.Unlock()

	// // built-in counters
	// st := GetSourceTime()
	// if st > 0 {
	// 	m[0] = uint64(st)
	// } else {
	// 	m[0] = 0
	// }

	// // here we add the information about sent data (counters)
	// m[TypeIndex] = CounterType
	return m
}

// func builtinCountersUpdate() {
// 	// Built-in counters update
// 	valmux.Lock()
// 	for id, name := range builtinCounterMap {
// 		switch name {
// 		case "PACKETS":
// 			counterMap[id].(*PACKETS).Increment()
// 		case "REAL_TIME":
// 			counterMap[id].(*REAL_TIME).Set(time.Now())
// 		case "SOURCE_TIME":
// 			counterMap[id].(*SOURCE_TIME).Set(SourceTime)
// 		}
// 	}
// 	valmux.Unlock()
// }

//----------------------------------------------------------------------------//
//------------------------------ MAIN FUNCTIONS ------------------------------//
//----------------------------------------------------------------------------//

// Sniff Open the device and start to sniff packets.
// These packets are sent to the the dispatcher
func Sniff() {
	// Open the handler according to the source (interface or file)
	if iface {
		handle, err = pcap.OpenLive(device,
			snapshotLen,
			promiscuous,
			pcap.BlockForever)
	} else {
		handle, err = pcap.OpenOffline(device)
	}
	if err != nil {
		minerLogger.Error().Msg(fmt.Sprintf("Error while opening device (%s)", err))
	}
	defer handle.Close()

	// init the number of parse packets
	var nbParsedPkts uint64
	// init the packet source from the handler
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	// Start all the counters (if they are not running)
	startAllCounters()
	// now we are sniffing!
	sniffing = true
	// pool of goroutines (for every packets)
	var goroutinePool sync.WaitGroup
	// create i/o channels
	initChannels()
	// loop over the incoming packets
	for {
		select {
		// events
		case e := <-defaultEventChannel:
			switch e {
			case STOP:
				// the counters are stopped
				minerLogger.Info().Msg("Stop sniffing")
				release(&goroutinePool, nil, nbParsedPkts)
				return
			case GET:
				// counter values are retrieved and sent
				// to the channel.
				minerLogger.Debug().Msg("Receiving GET")
				defaultDataChannel <- getValues(&goroutinePool)
			case FLUSH:
				// counter values are retrieved and sent
				// to the channel. The counters are also
				// reset.
				minerLogger.Debug().Msg("Receiving FLUSH")
				defaultDataChannel <- getValuesAndReset(&goroutinePool)
			default:
				minerLogger.Debug().Msgf("Receiving unknown event (%v)", e)
			}
		// parse a packet
		case packet, ok := <-packetSource.Packets():
			// check whether it is the last packet or not
			// fmt.Println(ok)
			if ok {
				// update time
				SourceTime = packet.Metadata().Timestamp
				// built-in counters update
				// builtinCountersUpdate()
				// add the packet parsing to the goroutine pool
				goroutinePool.Add(1)
				// dispatch the packet
				go dispatch(&goroutinePool, packet)
				// increment the number of parsed packets
				nbParsedPkts++

			} else {
				// if there is no packet anymore, we stop it
				minerLogger.Info().Msgf("No packets to parse anymore (%d parsed packets)",
					nbParsedPkts)
				minerLogger.Info().Msg("Stop sniffing")
				release(&goroutinePool, nil, nbParsedPkts)
				return
			}
		}

	}
}

// GoSniff sniffs in a detached goroutine. It returns an event channel to
// communicate with it and a data channel to retrieve the counter values
func GoSniff() (EventChannel, DataChannel) {
	// init the event channel
	eventChannel := make(EventChannel)
	// init the data channel (counters)
	dataChannel := make(DataChannel)

	go func() {
		// Open the handler according to the source (interface or file)
		if iface {
			handle, err = pcap.OpenLive(device,
				snapshotLen,
				promiscuous,
				pcap.BlockForever)
		} else {
			handle, err = pcap.OpenOffline(device)
		}
		if err != nil {
			minerLogger.Error().Msg(fmt.Sprintf("Error while opening device (%s)", err))
		}
		defer handle.Close()

		// init the number of parse packets
		var nbParsedPkts uint64
		// init the packet source from the handler
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		// Start all the counters (if they are not running)
		startAllCounters()
		// now we are sniffing!
		sniffing = true
		// pool of goroutines (for every packets)
		var goroutinePool sync.WaitGroup

		// loop over the incoming packets
		for {
			select {
			// events
			case e, _ := <-eventChannel:
				// minerLogger.Debug().
				// 	Str("Function", "GoSniff").
				// 	Int("Event", len(eventChannel)).
				// 	Int("Data", len(dataChannel)).
				// 	Msg("")
				switch e {
				case STOP:
					// the counters are stopped
					minerLogger.Info().Msg("Stop sniffing")
					release(&goroutinePool, dataChannel, nbParsedPkts)
					return
				case GET:
					// counter values are retrieved and sent
					// to the channel.
					minerLogger.Debug().Msg("Receiving GET")
					dataChannel <- getValues(&goroutinePool)
				case FLUSH:
					// counter values are retrieved and sent
					// to the channel. The counters are also
					// reset.
					goroutinePool.Wait()
					minerLogger.Debug().Msg("Receiving FLUSH")
					dataChannel <- getValuesAndReset(&goroutinePool)
				// case PERF:
				// 	// perf aims to send the current number of parsed
				// 	// packets
				// 	minerLogger.Debug().Msg("Receiving PERF")
				// 	dataChannel <- map[int]uint64{
				// 		TypeIndex: PerfType,
				// 		PerfIndex: nbParsedPkts,
				// 	}
				// case TIME:
				// 	// time aims to send the current source time
				// 	minerLogger.Debug().Msg("Receiving TIME")
				// 	dataChannel <- map[int]uint64{
				// 		TypeIndex: TimeType,
				// 		TimeIndex: uint64(SourceTime.UnixNano()),
				// 	}
				default:
					minerLogger.Debug().Msgf("Receiving unknown event (%v)", e)
				}
			// parse a packet
			case packet, ok := <-packetSource.Packets():
				// check whether it is the last packet or not
				if ok {
					// update time
					SourceTime = packet.Metadata().Timestamp
					// built-in counters update
					// builtinCountersUpdate()
					// add the packet parsing to the goroutine pool
					goroutinePool.Add(1)
					// dispatch the packet
					go dispatch(&goroutinePool, packet)
					// increment the number of parsed packets
					nbParsedPkts++

				} else {
					// if there is no packet anymore, we stop it
					minerLogger.Info().Msgf("No packets to parse anymore")
					release(&goroutinePool, dataChannel, nbParsedPkts)
					return
				}
			}

		}
	}()
	return eventChannel, dataChannel
}

// GoSniffAndYieldChannel sniffs in a detached goroutine. It returns an event channel to
// communicate with it and a data channel where counter values are sent through according
// to the sending period.
func GoSniffAndYieldChannel(period time.Duration) (EventChannel, DataChannel) {
	// init the event channel
	eventChannel := make(EventChannel)
	// init the data channel (counters)
	dataChannel := make(DataChannel)

	go func() {
		// Open the handler according to the source (interface or file)
		if iface {
			handle, err = pcap.OpenLive(device,
				snapshotLen,
				promiscuous,
				pcap.BlockForever)
		} else {
			handle, err = pcap.OpenOffline(device)
		}
		if err != nil {
			minerLogger.Error().Msg(fmt.Sprintf("Error while opening device (%s)", err))
		}
		defer handle.Close()

		// init the number of parse packets
		var nbParsedPkts uint64
		// init the packet source from the handler
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		// Start all the counters (if they are not running)
		startAllCounters()
		// now we are sniffing!
		sniffing = true
		// pool of goroutines (for every packets)
		var goroutinePool sync.WaitGroup
		// last timestamp (initialized with a forward time)
		lastTimestamp := time.Now().Add(1 * time.Hour)
		// bool to send data to the channel
		send := false

		// loop over the incoming packets
		for {
			select {
			// events
			case e, _ := <-eventChannel:
				// Aimed to debug channels
				// minerLogger.Debug().
				// 	Str("Function", "GoSniff").
				// 	Int("Event", len(eventChannel)).
				// 	Int("Data", len(dataChannel)).
				// 	Msg("")
				switch e {
				case STOP:
					// the counters are stopped
					minerLogger.Info().Msg("Stop sniffing")
					// close(eventChannel)
					release(&goroutinePool, dataChannel, nbParsedPkts)
					return
				case GET:
					// counter values are retrieved and sent
					// to the channel.
					minerLogger.Debug().Msg("Receiving GET")
					dataChannel <- getValues(&goroutinePool)
				case FLUSH:
					// counter values are retrieved and sent
					// to the channel. The counters are also
					// reset.
					minerLogger.Debug().Msg("Receiving FLUSH")
					dataChannel <- getValuesAndReset(&goroutinePool)
				// case PERF:
				// 	// perf aims to send the current number of parsed
				// 	// packets
				// 	minerLogger.Debug().Msg("Receiving PERF")
				// 	dataChannel <- map[int]uint64{
				// 		TypeIndex: PerfType,
				// 		PerfIndex: nbParsedPkts,
				// 	}
				// case TIME:
				// 	// time aims to send the current source time
				// 	minerLogger.Debug().Msg("Receiving TIME")
				// 	dataChannel <- map[int]uint64{
				// 		TypeIndex: TimeType,
				// 		TimeIndex: uint64(SourceTime.UnixNano()),
				// 	}
				default:
					minerLogger.Debug().Msgf("Receiving unknown event (%v)", e)
				}
			// parse a packet
			case packet, ok := <-packetSource.Packets():
				// check whether it is the last packet or not
				if ok {
					// update time and check if data must be sent through the channel
					if lastTimestamp, send = updateAndCheckTime(packet.Metadata().Timestamp,
						lastTimestamp,
						period); send {
						// send data to the channel (like FLUSH)
						dataChannel <- getValuesAndReset(&goroutinePool)
					}
					// add the packet parsing to the goroutine pool
					goroutinePool.Add(1)
					// dispatch the packet
					go dispatch(&goroutinePool, packet)
					// increment the number of parsed packets
					nbParsedPkts++
					// built-in counters update
					// builtinCountersUpdate()
				} else {
					// if there is no packet anymore, we stop it
					minerLogger.Info().Msgf("No packets to parse anymore (%d parsed packets)",
						nbParsedPkts)
					release(&goroutinePool, dataChannel, nbParsedPkts)
					close(dataChannel)
					return
				}
			}

		}
	}()
	return eventChannel, dataChannel
}

//----------------------------------------------------------------------------//
//--------------------------- UNEXPORTED FUNCTIONS ---------------------------//
//----------------------------------------------------------------------------//

// dispatch sends the incoming packet to the loaded
// counters
func dispatch(goroutinePool *sync.WaitGroup, pkt gopacket.Packet) {
	// NEW
	var arp *layers.ARP
	var ip *layers.IPv4
	var tcp *layers.TCP
	var udp *layers.UDP
	var icmp *layers.ICMPv4
	var p *counters.Pattern
	var ok bool

	ipLayer := pkt.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		if ip, ok = ipLayer.(*layers.IPv4); ok {
			p = counters.NewIPPattern()
			p.SetFromIPv4Layer(ip)
			switch ip.NextLayerType() {
			case layers.LayerTypeTCP:
				if tcp, ok = pkt.TransportLayer().(*layers.TCP); ok {
					p.SetFromTCPLayer(tcp)
				}
			case layers.LayerTypeUDP:
				if udp, ok = pkt.TransportLayer().(*layers.UDP); ok {
					p.SetFromUDPLayer(udp)
				}
			case layers.LayerTypeICMPv4:
				icmp, _ = pkt.Layer(layers.LayerTypeICMPv4).(*layers.ICMPv4)
			default:
				// do nothing
			}
		}
	} else {
		// NEW
		arp, _ = pkt.Layer(layers.LayerTypeARP).(*layers.ARP)
	}

	for _, ctr := range counterMap {
		switch ctr.(type) {
		// NEW
		case counters.ARPCtrInterface:
			if arp != nil {
				arpctr, _ := ctr.(counters.ARPCtrInterface)
				arpctr.LayPipe() <- arp
			}
		case counters.IPCtrInterface:
			if ip != nil {
				ipctr, _ := ctr.(counters.IPCtrInterface)
				ipctr.LayPipe() <- ip
			}
		case counters.TCPCtrInterface:
			if tcp != nil {
				tcpctr, _ := ctr.(counters.TCPCtrInterface)
				tcpctr.LayPipe() <- tcp
			}
		case counters.UDPCtrInterface:
			if udp != nil {
				udpctr, _ := ctr.(counters.UDPCtrInterface)
				udpctr.LayPipe() <- udp
			}
		case counters.ICMPCtrInterface:
			if icmp != nil {
				icmpctr, _ := ctr.(counters.ICMPCtrInterface)
				icmpctr.LayPipe() <- icmp
			}
		case counters.PatternCtrInterface:
			if p != nil {
				patternctr, _ := ctr.(counters.PatternCtrInterface)
				patternctr.LayPipe() <- p
			}
		case counters.PktCtrInterface:
			pktctr, _ := ctr.(counters.PktCtrInterface)
			pktctr.LayPipe() <- pkt
		}
	}
	goroutinePool.Done()
}

// func updateTime(t time.Time, send bool, tc TimeChannel) {
// 	SourceTime = t
// 	// minerLogger.Debug().Msgf("Updating time (time: %s, send ticks: %v)", t, sendTicks)
// 	if send {
// 		if SourceTime.Sub(last) < 0 {
// 			last = SourceTime
// 		} else if SourceTime.Sub(last) > tickPeriod {
// 			last = SourceTime
// 			minerLogger.Debug().Msg("Sending tick")
// 			if tc != nil {
// 				tc <- SourceTime
// 			} else {
// 				defaultTimeChannel <- SourceTime
// 			}
// 		}
// 	}
// }

func updateAndCheckTime(current time.Time, last time.Time, period time.Duration) (time.Time, bool) {
	// update the time (from the packet timestamp)
	SourceTime = current

	// if the last time stamp has not been initialized
	// returns the new last timestamp and the boolean value saying not
	// to send data
	if SourceTime.Sub(last) < 0 {
		return SourceTime, false
	}

	// send data and update last
	if SourceTime.Sub(last) > period {
		return SourceTime, true
	}

	// do nothing
	return last, false
}

// release changes the state of the package.
// func release(goroutinePool *sync.WaitGroup, tc TimeChannel) {
// goroutinePool.Wait()
// stopAllCounters()
// getValuesAndReset(nil)
// sniffing = false
// if sendTicks {
// 	// a last tick is send to trigger the analyzer tick event
// 	if tc != nil {
// 		tc <- time.Now()
// 	} else {
// 		defaultTimeChannel <- time.Now()
// 	}
// 	sendTicks = false
// }
// closeChannels()
// }

// release changes the state of the package.
func release(goroutinePool *sync.WaitGroup, data DataChannel, pkts uint64) {
	goroutinePool.Wait()
	// fmt.Println("STOP")
	stopAllCounters()
	minerLogger.Info().Msgf("Stop sniffing (%d parsed packets)", pkts)
	sniffing = false
	// send a nil value to say that nothing else will be sent
	// data <- nil
	// close(data)

}

func startAllCounters() error {
	minerLogger.Info().Msg("Starting all counters")
	for _, ctr := range counterMap {
		if !ctr.IsRunning() {
			go counters.Run(ctr)
		}
	}
	return nil
}

func startCounter(id int) error {
	ctr, ok := counterMap[id]
	if ok {
		if !ctr.IsRunning() {
			ctr.SwitchRunningOn()
			go counters.Run(ctr)
		} else {
			return fmt.Errorf("The counter %s is already running", ctr.Name())
		}
	} else {
		return fmt.Errorf("The ID %d does not refer to a loaded counter", id)
	}
	return nil
}

func stillRunningCounters() bool {
	for _, ctr := range counterMap {
		if ctr.IsRunning() {
			return true
		}
	}
	return false
}

func stopAllCounters() error {
	for _, ctr := range counterMap {
		if ctr.IsRunning() {
			ctr.SigPipe() <- 0
		}
	}
	minerLogger.Info().Msg("Stopping all counters")
	// to be sure they are all stopped
	for stillRunningCounters() {
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}
