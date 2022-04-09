package miner

import (
	"time"

	"github.com/google/gopacket"
)

// sniffOffline opens an interface and starts to sniff.
// It sends counters snapshot at given period
func sniffOffline(packetChan chan gopacket.Packet,
	period time.Duration,
	data DataChannel) error {
	// now we are sniffing!
	minerLogger.Debug().Msgf("Sniffing file...")
	sniffing.Begin()
	// set running to false when exits
	defer release()

	// Treat the first packet
	firstPacket := <-packetChan
	dispatcher.dispatch(firstPacket)
	// init the first timestamp
	lastTick := firstPacket.Metadata().Timestamp

	// loop over the incoming packets
	for {
		select {
		// manage events
		case e := <-internalEventChannel:
			switch e {
			case STOP:
				// the counters are stopped
				minerLogger.Debug().Msg("Receiving STOP")
				dispatcher.terminate()
				minerLogger.Debug().Msg("Dispatcher has terminated")
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
			dispatcher.dispatch(packet)

			// update the timestamp
			sourceTime.Set(packet.Metadata().Timestamp)

			// send data at given period
			st := sourceTime.Get()
			if st.Sub(lastTick) > period {
				lastTick = st
				data <- dispatcher.terminateAndFlushAll()
			}
		}

	}

}
