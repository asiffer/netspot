// // pattern.go

// package counters

// import (
// 	"fmt"
// 	"net"
// 	"strconv"
// 	"strings"
// 	"sync/atomic"

// 	"github.com/google/gopacket/layers"
// )

// var (
// 	// NullIPNet is the undefined IP
// 	NullIPNet = &net.IPNet{
// 		IP:   net.IP{0, 0, 0, 0},
// 		Mask: net.IPMask{0, 0, 0, 0},
// 	}
// 	// NullPort is the undefined port
// 	NullPort           = 0
// 	endpointsSeparator = "->"
// 	authorizedRunes    = "0123456789.:-/>abcdef"
// 	maxPort            = 65535
// )

// // Pattern defines a basic flow src:srcPort -> dst:dstPort
// type Pattern struct {
// 	src     *net.IPNet
// 	dst     *net.IPNet
// 	srcPort int
// 	dstPort int
// }

// // NewIPPattern returns a new pattern aimed to only define an
// // IP (not a subnetwork). So its mask is set to 32.
// func NewIPPattern() *Pattern {
// 	return &Pattern{
// 		src: &net.IPNet{
// 			IP:   net.IP{0, 0, 0, 0},
// 			Mask: net.IPMask{255, 255, 255, 255},
// 		},
// 		dst: &net.IPNet{
// 			IP:   net.IP{0, 0, 0, 0},
// 			Mask: net.IPMask{255, 255, 255, 255},
// 		},
// 		srcPort: 0,
// 		dstPort: 0,
// 	}
// }

// func parseEndpoint(e string) (*net.IPNet, int, error) {
// 	var subnet *net.IPNet
// 	var port int
// 	var err error
// 	if !strings.ContainsRune(e, ':') {
// 		return nil, -1, fmt.Errorf("Character ':' not found in %s", e)
// 	}
// 	s := strings.Split(e, ":")
// 	// s[0] contains subnet
// 	// s[1] contains port

// 	if len(s[0]) == 0 {
// 		subnet = NullIPNet
// 	} else if strings.ContainsRune(s[0], '/') {
// 		addr := strings.Split(s[0], "/")

// 		ip := net.ParseIP(addr[0])
// 		if ip == nil {
// 			return nil, -1, fmt.Errorf("Error while parsing %s", addr[0])
// 		}
// 		_, maxMaskSize := ip.DefaultMask().Size()

// 		msize, err := strconv.Atoi(addr[1])
// 		if err != nil {
// 			return nil, -1, err
// 		} else if msize > maxMaskSize || msize < 0 {
// 			return nil, -1, fmt.Errorf("Mask size must be between 0 and %d", maxMaskSize)
// 		}

// 		subnet = &net.IPNet{
// 			IP:   ip,
// 			Mask: net.CIDRMask(msize, maxMaskSize),
// 		}
// 	} else {
// 		ip := net.ParseIP(s[0])
// 		if ip == nil {
// 			return nil, -1, fmt.Errorf("Error while parsing %s", s[0])
// 		}
// 		_, maxMaskSize := ip.DefaultMask().Size()
// 		fullMask := net.CIDRMask(maxMaskSize, maxMaskSize)

// 		subnet = &net.IPNet{
// 			IP:   ip,
// 			Mask: fullMask,
// 		}
// 	}

// 	if len(s[1]) > 0 { // if a port is defined
// 		port, err = strconv.Atoi(s[1])
// 		if err != nil {
// 			return nil, -1, err
// 		} else if port < 0 || port > maxPort {
// 			return nil, -1, fmt.Errorf("Port must be between 0 and %d", maxPort)
// 		}
// 	} else { // otherwise we accept all the ports
// 		port = 0
// 	}

// 	return subnet, port, nil
// }

// func clean(s string) string {
// 	mapping := func(r rune) rune {
// 		if strings.ContainsRune(authorizedRunes, r) {
// 			return r // keep the value
// 		}
// 		return -1 // remove the value
// 	}
// 	return strings.Map(mapping, s)
// }

// // ParsePattern tries to parse a pattern from a string
// // with the model ip/mask:port->ip/mask:port
// // ('->' and ':' are mandatory)
// func ParsePattern(s string) (*Pattern, error) {
// 	s = clean(s) // remove unknown characters
// 	ep := strings.Split(s, endpointsSeparator)
// 	if len(ep) != 2 {
// 		return nil, fmt.Errorf("Separator %s not found",
// 			endpointsSeparator)
// 	}

// 	src, srcPort, err := parseEndpoint(ep[0])
// 	if err != nil {
// 		return nil, err
// 	}

// 	dst, dstPort, err := parseEndpoint(ep[1])
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Pattern{
// 		src:     src,
// 		dst:     dst,
// 		srcPort: srcPort,
// 		dstPort: dstPort,
// 	}, nil
// }

// func (p *Pattern) String() string {
// 	return fmt.Sprintf("%v:%d->%v:%d",
// 		p.src,
// 		p.srcPort,
// 		p.dst,
// 		p.dstPort)
// }

// // ToBPF maps the pattern to a BPF expression
// func (p *Pattern) ToBPF() string {
// 	bpfList := make([]string, 0)
// 	ones := -1
// 	if ones, _ = p.src.Mask.Size(); ones != 0 {
// 		bpfList = append(bpfList, fmt.Sprintf("src net %v", p.src))
// 	}
// 	if ones, _ = p.dst.Mask.Size(); ones != 0 {
// 		bpfList = append(bpfList, fmt.Sprintf("dst net %v", p.dst))
// 	}
// 	if p.srcPort > 0 {
// 		bpfList = append(bpfList, fmt.Sprintf("src port %d", p.srcPort))
// 	}
// 	if p.dstPort > 0 {
// 		bpfList = append(bpfList, fmt.Sprintf("dst port %d", p.dstPort))
// 	}
// 	return strings.Join(bpfList, " and ")
// }

// // SetFromIPv4Layer changes the IP addresses of the Pattern according to
// // the given IP layer
// func (p *Pattern) SetFromIPv4Layer(layer *layers.IPv4) {
// 	p.src.IP = layer.SrcIP
// 	p.dst.IP = layer.DstIP
// }

// // SetFromTCPLayer changes the IP addresses of the Pattern according to
// // the given TCP layer
// func (p *Pattern) SetFromTCPLayer(layer *layers.TCP) {
// 	p.srcPort = int(layer.SrcPort)
// 	p.dstPort = int(layer.DstPort)
// }

// // SetFromUDPLayer changes the IP addresses of the Pattern according to
// // the given UDP layer
// func (p *Pattern) SetFromUDPLayer(layer *layers.UDP) {
// 	p.srcPort = int(layer.SrcPort)
// 	p.dstPort = int(layer.DstPort)
// }

// func portsMatch(p, other int) bool {
// 	return (p == other) || (p == NullPort)
// }

// // Contains checks if a pattern (p) include another pattern (other)
// func (p *Pattern) Contains(other *Pattern) bool {
// 	return p.src.Contains(other.src.IP) &&
// 		p.dst.Contains(other.dst.IP) &&
// 		portsMatch(p.srcPort, other.srcPort) &&
// 		portsMatch(p.dstPort, other.dstPort)
// }

// // PatternCtrInterface is the interface defining a network pattern
// // The paramount method is obviously 'process'
// type PatternCtrInterface interface {
// 	BaseCtrInterface
// 	Process(*Pattern) // method to process a packet
// 	LayPipe() chan *Pattern
// 	// ToBPF() string
// }

// // PatternCtr is the pattern counter
// type PatternCtr struct {
// 	BaseCtr
// 	pattern *Pattern
// 	name    string
// 	Counter uint64
// 	Lay     chan *Pattern
// }

// // Name returns the name of the pattern (namely a tag)
// func (ctr *PatternCtr) Name() string {
// 	return ctr.name
// }

// // Process update the counter according to data it receives
// func (ctr *PatternCtr) Process(p *Pattern) {
// 	if ctr.pattern.Contains(p) {
// 		atomic.AddUint64(&ctr.Counter, 1)
// 	}
// }

// // Increment only increment the underlying counter (aim to
// // work with BPF expressions)
// func (ctr *PatternCtr) Increment() {
// 	atomic.AddUint64(&ctr.Counter, 1)
// }

// // Reset resets the counter
// func (ctr *PatternCtr) Reset() {
// 	atomic.StoreUint64(&ctr.Counter, 0)
// }

// // Value returns the current value of the counter (method of BaseCtrInterface)
// func (ctr *PatternCtr) Value() uint64 {
// 	return atomic.LoadUint64(&ctr.Counter)
// }

// // NewPatternCtr is the generic constructor of a pattern counter
// func NewPatternCtr(p *Pattern, name string) *PatternCtr {
// 	return &PatternCtr{
// 		pattern: p,
// 		name:    name,
// 		Counter: 0,
// 		Lay:     make(chan *Pattern)}
// }

// // LayPipe returns the flow channel of the pattern counter
// func (ctr *PatternCtr) LayPipe() chan *Pattern {
// 	return ctr.Lay
// }

// // RunPatternCtr starts a pattern counter
// func RunPatternCtr(ctr PatternCtrInterface) {
// 	ctr.SwitchRunningOn()
// 	for {
// 		select {
// 		case sig := <-ctr.SigPipe():
// 			switch sig {
// 			case STOP: // stop the counter
// 				ctr.SwitchRunningOff()
// 				return
// 			case GET: // return the value
// 				ctr.ValPipe() <- ctr.Value()
// 			case RESET: // reset
// 				ctr.Reset()
// 			case FLUSH: // return the value and reset
// 				ctr.ValPipe() <- ctr.Value()
// 				ctr.Reset()
// 			}
// 		case flow := <-ctr.LayPipe(): // process the packet
// 			ctr.Process(flow)

// 		}
// 	}
// }
