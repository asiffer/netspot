package collector

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/asiffer/netspot/collector/xdp"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

const OPERATION = "op"

const MAX_UINT16 = uint32(65536)

func sum(a []uint64) uint64 {
	var s uint64 = 0
	for _, v := range a {
		s += v
	}
	return s
}

// XDPRread populates a Data structure from the Map defined
// in hook.c
func XDPReadCounters(dst *Data, src *ebpf.Map) error {
	// the maps is per cpu so cilium returns an array of values for each key,
	// we need to sum them up
	counts := make([]uint64, 0)
	for _, k := range Keys() {
		err := src.Lookup(k, &counts)
		if err != nil {
			return err
		} else {
			*dst.Addr(k) = sum(counts)
		}
	}
	return nil
}

func XDPReadDstPorts(dst *Data, src *ebpf.Map) error {
	counts := make([]uint64, 0)
	for i := range MAX_UINT16 {
		err := src.Lookup(i, &counts)
		if err != nil {
			return err
		} else {
			dst.TCP_DST_PORTS[i] = sum(counts)
		}
	}
	return nil
}

// XDPCollector is an eBPF-based data collector. It leverages XDP to grab
// network counters on interfaces before any kernel processing.
type XDPCollector struct {
	CollectorHooks
	// xdp objects
	objs xdp.XDPObjects
	link link.Link
	// config
	iface *net.Interface
	tick  time.Duration
}

// NewXDPCollector creates a new collector given the network interface and
// the tick time
func NewXDPCollector(iface *net.Interface, tick time.Duration) Collector {
	return &XDPCollector{
		CollectorHooks: NewCollectorHooks(),
		iface:          iface,
		tick:           tick,
	}
}

func (c *XDPCollector) Config() CollectorConfig {
	return CollectorConfig{
		Source: c.iface.Name,
		Tick:   c.tick,
	}
}

// Load loads ebpf object and attaches it to configured interface
func (c *XDPCollector) Load() error {
	var err error
	c.log("remove memory lock")
	if err = rlimit.RemoveMemlock(); err != nil {
		return err
	}

	c.log("load eBPF objects")
	if err = xdp.LoadXDPObjects(&c.objs, nil); err != nil {
		return err
	}

	c.log(fmt.Sprintf("attach program to XDP hook (%s)", c.iface.Name))
	if c.link, err = link.AttachXDP(link.XDPOptions{
		Program:   c.objs.XdpUpdateCounters,
		Interface: c.iface.Index,
	}); err != nil {
		return err
	}

	return nil
}

// Unload detaches the program from the interface and unloads the ebpf object
func (c *XDPCollector) Unload() error {
	// unattach
	c.log(fmt.Sprintf("detach program from XDP hook (%s)", c.iface.Name))
	if err := c.link.Close(); err != nil {
		return err
	}
	// unload objects
	c.log("unload eBPF objects")
	if err := c.objs.Close(); err != nil {
		return err
	}

	return nil
}

// Start creates a goroutine that periodically fetches data from the
// XDP hook
func (c *XDPCollector) Start(stop chan bool) {
	tick := time.NewTicker(c.tick)
	out := Data{}
	for {
		select {
		case t := <-tick.C:
			// prepare data
			out.TIME = uint64(t.UnixNano())
			// *out = Data{TIME: uint64(t.UnixNano())}
			// populate
			err := errors.Join(
				XDPReadCounters(&out, c.objs.NetspotCountersMap),
				XDPReadDstPorts(&out, c.objs.NetspotTcpDstPortsMap),
			)

			// check for errors
			if err != nil {
				c.err(err)
			} else {
				// otherwise dispatch data
				c.send(&out)
			}
		case <-stop:
			c.log("receiving stop signal")
			c.end(nil)
			return
		}
	}
}
