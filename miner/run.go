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
	"github.com/rs/zerolog/log"
)

const (
	// reserved keys for data different from
	// counter values (index of counters start
	// from 0)
	perfIndex int = -1
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

// GetNbParsedPkts returns the number of current parsed packets since
// the initialization
// func GetNbParsedPkts() uint64 {
// return nbParsedPkts
// }

// // StartSniffing starts to sniff the current device. It does nothing
// // if the sniffing is in progress. This is a goroutine, so it returns
// // once the sniffing has started.
// func StartSniffing() {
// 	if !sniffing {
// 		log.Info().Msgf("Start sniffing %s", device)
// 		log.Debug().Msg(fmt.Sprint("Loaded counters: ", GetLoadedCounters()))
// 		events = make(chan uint8)
// 		go Sniff()
// 		for !sniffing {
// 			// wait for sniffing
// 		}
// 	}

// }

// StartSniffing starts to sniff the current device. It does nothing
// if the sniffing is in progress. This is a goroutine, so it returns
// once the sniffing has started.
func StartSniffing() (EventChannel, DataChannel, TimeChannel) {
	if !sniffing {
		log.Info().Msgf("Start sniffing %s", device)
		log.Debug().Msg(fmt.Sprint("Loaded counters: ", GetLoadedCounters()))
		// events = make(chan uint8)
		ec, dc, tc := GoSniff()
		for !sniffing {
			// wait for sniffing
		}
		// update the default channels
		defaultEventChannel = ec
		defaultDataChannel = dc
		defaultTimeChannel = tc
		return ec, dc, tc
	}
	log.Debug().Msg("StartSniffing requested but already sniffing")
	return defaultEventChannel, defaultDataChannel, defaultTimeChannel
}

// StartSniffingAndWait sniff the current device but does not
// return immediately. This is more relevant when the device
// is a capture file: the sniff stops when all the packets have
// been read.
func StartSniffingAndWait() {
	if !sniffing {
		log.Info().Msgf("Start sniffing %s", device)
		log.Debug().Msg(fmt.Sprint("Loaded counters: ", GetLoadedCounters()))
		// events = make(chan uint8)
		Sniff()
	}
	log.Info().Msg("Sniffing stopped")
}

// SniffAndSend start sniffing and returns periodically
// counter values
func SniffAndSend(d time.Duration, reset bool, socket string) {
	// connect to the socket
	conn, err := net.Dial("unixgram", socket)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	defer conn.Close()
	// initialize the data encoder. It will directly
	// sned the data to the socket
	encoder := gob.NewEncoder(conn)
	// init the ticker
	internalTicker := time.Tick(d)
	// start the sniffing process (get the channels)
	ec, dc, _ := GoSniff()

	// loop
	for {
		select {
		// wait for a tick event to send a snapshot
		case <-internalTicker:
			log.Debug().Msg("Tick")
			if !sniffing {
				// trigger the snapshot, and get data
				m := SnapshotID(reset, ec, dc)
				log.Debug().Msgf("%v", m)
				// encode and send data
				encoder.Encode(m)
				// close the connection
				log.Warn().Msg("Sending EOF")
				conn.Write([]byte{'E', 'O', 'F'})
				return
			}
			// trigger the snapshot, and get data
			m := SnapshotID(reset, ec, dc)
			log.Debug().Msgf("%v", m)
			// encode and send data
			encoder.Encode(m)
		}
	}
}

// StopSniffing stops to sniff the device
func StopSniffing(ec EventChannel) {
	if sniffing {
		if ec != nil {
			ec <- STOP
		} else {
			defaultEventChannel <- STOP
		}
	}
}

// GetCounterValue returns the current value of the counter
// identified by its id
func GetCounterValue(id int) (uint64, error) {
	ctr, ok := counterMap[id]
	if !ok {
		log.Error().Msg("Invalid counter identifier")
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
		log.Error().Msg("Invalid counter identifier")
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
			log.Error().Msg(err.Error())
		}
	}
	valmux.Unlock()
	st := GetSourceTime()
	if st > 0 {
		m[0] = uint64(st)
	} else {
		m[0] = 0
	}
	return m
}

// func getValuesAndReset(pool *sync.WaitGroup) {
// 	if pool != nil {
// 		pool.Wait()
// 	}
// 	valmux.Lock()
// 	for i := range counterMap {
// 		if v, err := GetCounterValueAndReset(i); err == nil {
// 			counterValues[i] = v
// 		} else {
// 			log.Error().Msg(err.Error())
// 		}
// 	}
// 	valmux.Unlock()
// }

func getValuesAndReset(pool *sync.WaitGroup) map[int]uint64 {
	m := make(map[int]uint64)
	if pool != nil {
		pool.Wait()
	}
	valmux.Lock()
	for i := range counterMap {
		if v, err := GetCounterValueAndReset(i); err == nil {
			// counterValues[i] = v
			m[i] = v
		} else {
			log.Error().Msg(err.Error())
		}
	}
	valmux.Unlock()
	st := GetSourceTime()
	if st > 0 {
		m[0] = uint64(st)
	} else {
		m[0] = 0
	}
	return m
}

// GetNbParsedPkts returns the current number of parsed packets
func GetNbParsedPkts() uint64 {
	// func GetPerf() uint64 {
	if sniffing {
		fmt.Println("NbParsedPAckets: sniffing")
		defaultEventChannel <- PERF
		n := <-defaultDataChannel
		return n[perfIndex]
	}
	return nbParsedPkts
}

// Snapshot retrieve the current value of the counters. Their
// values are then put in the counterValues container. If nil is
// passed for a channel, then the default channel is used.
func Snapshot(reset bool, eChannel chan uint8, dChannel chan map[int]uint64) map[string]uint64 {
	m := SnapshotID(reset, eChannel, dChannel)
	s := make(map[string]uint64)
	for i, v := range m {
		// 0 is the time
		if i > 0 {
			s[counterMap[i].Name()] = v
		}
	}
	return s
}

// // Snapshot retrieve the current value of the counters. Their
// // values are then put in the counterValues container.
// func Snapshot(reset bool) map[string]uint64 {
// 	e := -1
// 	if reset {
// 		e = 2
// 	} else {
// 		e = 1
// 	}

// 	m := make(map[string]uint64)
// 	st := GetSourceTime()

// 	if sniffing {
// 		// trigger "getValues"
// 		events <- e
// 		// get a more accurate time
// 		st = GetSourceTime()
// 		valmux.Lock()
// 		for i, v := range counterValues {
// 			m[counterMap[i].Name()] = v
// 		}
// 		valmux.Unlock()
// 	} else {
// 		// when it is not sniffing we need to reset manually
// 		// (without sending the signal)
// 		if reset {
// 			getValuesAndReset(nil)
// 		} else {
// 			getValues(nil)
// 		}
// 		// then we can retrieve the values
// 		for i, v := range counterValues {
// 			m[counterMap[i].Name()] = v
// 		}

// 	}

// 	// check time
// 	if st > 0 {
// 		m["time"] = uint64(st)
// 	} else {
// 		m["time"] = 0
// 	}

// 	return m
// }

// SnapshotID retrieve the current value of the counters. Their
// values are then put in the counterValues container. If nil is
// passed for a channel, then the default channel is used.
func SnapshotID(reset bool, eChannel chan uint8, dChannel chan map[int]uint64) map[int]uint64 {
	// if the channel is not given, we use the default channel
	if dChannel == nil {
		dChannel = defaultDataChannel
	}
	if eChannel == nil {
		eChannel = defaultEventChannel
	}
	var e uint8
	// send the right signal to send
	if reset {
		e = FLUSH
	} else {
		e = GET
	}

	if sniffing {
		log.Debug().Msgf("Sending event %v", e)
		eChannel <- e
		return <-dChannel
	} else if reset {
		return getValuesAndReset(nil)
	} else {
		return getValues(nil)
	}

}

// Tick returns a channel sending frequent tick (at each duration)
// func Tick(d time.Duration) chan time.Time {
// 	last = SourceTime.Add(6 * time.Hour)
// 	tickPeriod = d
// 	sendTicks = true
// 	defaultTicker = make(chan time.Time)
// 	return defaultTicker
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
		log.Error().Msg(fmt.Sprintf("Error while opening device (%s)", err))
	}
	defer handle.Close()

	// reinit the number of parse packets
	nbParsedPkts = 0
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
				log.Info().Msg("Stop sniffing")
				release(&goroutinePool, nil)
				return
			case GET:
				// counter values are retrieved and sent
				// to the channel.
				log.Debug().Msg("Receiving GET")
				defaultDataChannel <- getValues(&goroutinePool)
			case FLUSH:
				// counter values are retrieved and sent
				// to the channel. The counters are also
				// reset.
				log.Debug().Msg("Receiving FLUSH")
				defaultDataChannel <- getValuesAndReset(&goroutinePool)
			case PERF:
				// perf aims to send the current number of parsed
				// packets
				log.Debug().Msg("Receiving PERF")
				defaultDataChannel <- map[int]uint64{perfIndex: nbParsedPkts}
			default:
				log.Debug().Msgf("Receiving unknown event (%v)", e)
			}
		// parse a packet
		case packet, ok := <-packetSource.Packets():
			// check whether it is the last packet or not
			if ok {
				updateTime(packet.Metadata().Timestamp, sendTicks, nil)
				// add the packet parsing to the goroutine pool
				goroutinePool.Add(1)
				// dispatch the packet
				go dispatch(&goroutinePool, packet)
				// increment the number of parsed packets
				nbParsedPkts++
			} else {
				// if there is no packet anymore, we stop it
				log.Info().Msgf("No packets to parse anymore (%d parsed packets)",
					nbParsedPkts)
				log.Info().Msg("Stop sniffing")
				release(&goroutinePool, nil)
				return
			}
		}

	}
}

// GoSniff sniffs in a detached goroutine. It returns the event channel to
// comunicate with it.
func GoSniff() (EventChannel, DataChannel, TimeChannel) {
	// init the event channel
	eChannel := make(EventChannel)
	// init the data channel (counters)
	dChannel := make(DataChannel)
	// init the time channel (send ticks)
	tChannel := make(TimeChannel)

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
			log.Error().Msg(fmt.Sprintf("Error while opening device (%s)", err))
		}
		defer handle.Close()

		// reinit the number of parse packets
		nbParsedPkts = 0
		// init the packet source from the handler
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		// Start all the counters (if they are not running)
		startAllCounters()
		// now we are sniffing!
		sniffing = true
		// pool of goroutines (for every packets)
		var goroutinePool sync.WaitGroup

		nt := 0
		// loop over the incoming packets

		for {
			select {
			// events
			case e := <-eChannel:
				switch e {
				case STOP:
					// the counters are stopped
					log.Info().Msg("Stop sniffing")
					release(&goroutinePool, tChannel)
					return
				case GET:
					// counter values are retrieved and sent
					// to the channel.
					log.Debug().Msg("Receiving GET")
					dChannel <- getValues(&goroutinePool)
				case FLUSH:
					// counter values are retrieved and sent
					// to the channel. The counters are also
					// reset.
					log.Debug().Msg("Receiving FLUSH")
					dChannel <- getValuesAndReset(&goroutinePool)
				case PERF:
					// perf aims to send the current number of parsed
					// packets
					log.Debug().Msg("Receiving PERF")
					defaultDataChannel <- map[int]uint64{perfIndex: nbParsedPkts}
				default:
					log.Debug().Msgf("Receiving unknown event (%v)", e)
				}
			// parse a packet
			case packet, ok := <-packetSource.Packets():
				// check whether it is the last packet or not
				if ok {
					SourceTime = packet.Metadata().Timestamp
					// log.Debug().Msgf("Updating time (time: %s, send ticks: %v)", t, sendTicks)
					if sendTicks {
						if SourceTime.Sub(last) < 0 {
							last = SourceTime
						} else if SourceTime.Sub(last) > tickPeriod {
							nt++
							last = SourceTime
							log.Debug().Msg("Sending tick")
							tChannel <- packet.Metadata().Timestamp
						}
					}
					// updateTime(packet.Metadata().Timestamp, sendTicks, tChannel)
					// add the packet parsing to the goroutine pool
					goroutinePool.Add(1)
					// dispatch the packet
					go dispatch(&goroutinePool, packet)
					// increment the number of parsed packets
					nbParsedPkts++
				} else {
					// if there is no packet anymore, we stop it
					log.Info().Msgf("No packets to parse anymore (%d parsed packets)",
						nbParsedPkts)
					log.Info().Msgf("%d ticks sent", nt)
					log.Info().Msg("Stop sniffing")
					release(&goroutinePool, tChannel)
					return
				}
			}

		}
	}()
	return eChannel, dChannel, tChannel
}

//----------------------------------------------------------------------------//
//--------------------------- UNEXPORTED FUNCTIONS ---------------------------//
//----------------------------------------------------------------------------//

// dispatch sends the incoming packet to the loaded
// counters
func dispatch(goroutinePool *sync.WaitGroup, pkt gopacket.Packet) {
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
	}

	for _, ctr := range counterMap {
		switch ctr.(type) {
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
		}
	}
	goroutinePool.Done()
}

func updateTime(t time.Time, send bool, tc TimeChannel) {
	SourceTime = t
	// log.Debug().Msgf("Updating time (time: %s, send ticks: %v)", t, sendTicks)
	if send {
		if SourceTime.Sub(last) < 0 {
			last = SourceTime
		} else if SourceTime.Sub(last) > tickPeriod {
			last = SourceTime
			log.Debug().Msg("Sending tick")
			if tc != nil {
				tc <- SourceTime
			} else {
				defaultTimeChannel <- SourceTime
			}
		}
	}

}

// release changes the state of the package.
func release(goroutinePool *sync.WaitGroup, tc TimeChannel) {
	goroutinePool.Wait()
	stopAllCounters()
	// getValuesAndReset(nil)
	sniffing = false
	if sendTicks {
		// a last tick is send to trigger the analyzer tick event
		if tc != nil {
			tc <- time.Now()
		} else {
			defaultTimeChannel <- time.Now()
		}
		sendTicks = false
	}
	closeChannels()
}

func startAllCounters() error {
	log.Info().Msg("Starting all counters")
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
	log.Info().Msg("Stopping all counters")
	// to be sure they are all stopped
	for stillRunningCounters() {
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}
