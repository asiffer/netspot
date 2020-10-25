package miner

import (
	"time"

	"github.com/google/gopacket"
)

// TODO: Maybe the support of PF_RING must be given up
// since it requires kernel modules. I would prefer a
// standalone binary first (full static with musl and libpcap)
// The support of PF_RING can be optional according to the
// user system.
// I should look at the AD_XDP socket.

// if iface {
// in case of network interface
// use PF_RING
// flag := pfring.FlagTimestamp
// if promiscuous {
// 	flag = flag | pfring.FlagPromisc
// }
// ringHandle, err = pfring.NewRing(device, uint32(snapshotLen), flag)
// if err != nil {
// 	return nil, err
// }

// PF_RING has a ton of optimizations and tweaks to
// make sure you get just the packets you want. For
// example, if you're only using pfring to read packets,
// consider running:
// ringHandle.SetSocketMode(pfring.ReadOnly)

// If you only care about packets received on your
// interface (not those transmitted by the interface),
// you can run:
// ringHandle.SetDirection(pfring.ReceiveOnly)

// if err = ringHandle.Enable(); err != nil { // Must do this!, or you get no packets!
// 	return nil, err
// }
// return gopacket.NewPacketSource(ringHandle, layers.LinkTypeEthernet), nil
// handle, err = pcap.OpenLive(device, snapshotLen, promiscuous, 1*time.Millisecond)
// if err != nil {
// 	minerLogger.Error().Msgf("Error while opening device: %v", err)
// 	return nil, err
// }
// return gopacket.NewPacketSource(handle, handle.LinkType()), nil
// }

// Otherwise use libpcap
// handle, err = pcap.OpenOffline(device)
// if err != nil {
// 	minerLogger.Error().Msgf("Error while opening device: %v", err)
// 	return nil, err
// }

// // init the packet source from the handler
// return gopacket.NewPacketSource(handle, handle.LinkType()), nil
// }

//TODO: In live case, if no packet arrives on the interface, the time
// is not updated, so sometimes no logs are flushed.

// sniffOnline opens an interface and starts to sniff.
// It sends counters snapshot at given period
func sniffOnline(packetChan chan gopacket.Packet, period time.Duration, event EventChannel, data DataChannel) error {

	// set the flush tick
	timeSource := time.Tick(period)

	// now we are sniffing!
	sniffing = true
	minerLogger.Debug().Msgf("Sniffing...")

	// loop over the incoming packets
	for {
		select {
		// periodic flush
		case SourceTime = <-timeSource:
			dispatcher.terminate()
			data <- dispatcher.flushAll()
		// manage events
		case e, _ := <-event:
			switch e {
			case STOP:
				// the counters are stopped
				minerLogger.Debug().Msg("Receiving STOP")
				dispatcher.terminate()
				sniffing = false
				return nil
			default:
				minerLogger.Debug().Msgf("Receiving unknown event (%v)", e)
			}
		// parse packet
		case packet, ok := <-packetChan:
			// check whether it is the last packet or not
			// if there is no packet anymore, we stop it
			if !ok {
				minerLogger.Info().Msgf("No packets to parse anymore (%d parsed packets).",
					dispatcher.receivedPackets)
				dispatcher.terminate()
				sniffing = false
				return nil
			}

			// in real packet case, dispatch the packet to the counters
			dispatcher.pool.Add(1)
			go dispatcher.dissect(packet)
		}

	}

}
