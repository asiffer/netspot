package miner

import (
	"time"

	"github.com/google/gopacket"
)

// sniffOffline opens an interface and starts to sniff.
// It sends counters snapshot at given period
func sniffOffline(packetChan chan gopacket.Packet, period time.Duration, event EventChannel, data DataChannel) error {
	// now we are sniffing!
	minerLogger.Debug().Msgf("Sniffing...")
	sniffing = true
	// set running to false when exits
	defer func() { sniffing = false }()

	// Treat the first packet
	pkt := <-packetChan
	dispatcher.pool.Add(1)
	go dispatcher.dissect(pkt)
	// init the first timestamp
	lastTick := pkt.Metadata().Timestamp

	// loop over the incoming packets
	for {
		select {
		// manage events
		case e, _ := <-event:
			switch e {
			case STOP:
				// the counters are stopped
				minerLogger.Debug().Msg("Receiving STOP")
				dispatcher.terminate()
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
				return nil
			}

			// in real packet case, dispatch the packet to the counters
			dispatcher.pool.Add(1)
			go dispatcher.dissect(packet)

			// update the timestamp
			SourceTime = packet.Metadata().Timestamp
			// send data at given period
			if SourceTime.Sub(lastTick) > period {
				lastTick = SourceTime
				dispatcher.terminate()
				data <- dispatcher.flushAll()
			}

		}

	}

}
