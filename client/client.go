package main

import (
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"log"
	"net"
	tools "tcp_server/toos"
	"time"
)

type client struct {
	// iface is the interface to send packets on.
	iface *net.Interface
	// destination, gateway (if applicable), and source IP addresses to use.
	dst, gw, src net.IP
	fakeSrc      net.IP

	handle *pcap.Handle

	// opts and buf allow us to easily serialize packets in the send()
	// method.
	opts gopacket.SerializeOptions
	buf  gopacket.SerializeBuffer
}

func NewClient(dst, fakeSrc net.IP, router routing.Router) (*client, error) {
	s := &client{
		dst: dst,
		opts: gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		buf: gopacket.NewSerializeBuffer(),
	}
	// Figure out the route to the IP.
	iface, gw, src, err := router.Route(dst)
	if err != nil {
		return nil, err
	}
	log.Printf("dst ip %v , interface %v, gateway %v, src %v\n", dst, iface.Name, gw, src)
	s.gw, s.src, s.iface = gw, src, iface

	handle, err := pcap.OpenLive(iface.Name,
		65536, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	s.handle = handle
	s.fakeSrc = fakeSrc
	return s, nil
}

// close cleans up the handle.
func (s *client) Close() {
	s.handle.Close()
}

// getHwAddr is a hacky but effective way to get the destination hardware
// address for our packets.  It does an ARP request for our gateway (if there is
// one) or destination IP (if no gateway is necessary), then waits for an ARP
// reply.  This is pretty slow right now, since it blocks on the ARP
// request/reply.
func (s *client) getHwAddr() (net.HardwareAddr, error) {
	start := time.Now()
	arpDst := s.dst
	if s.gw != nil {
		arpDst = s.gw
	}
	// Prepare the layers to send for an ARP request.
	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(s.iface.HardwareAddr),
		SourceProtAddress: []byte(s.src),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(arpDst),
	}
	// Send a single ARP request packet (we never retry a send, since this
	// is just an example ;)
	if err := s.send(&eth, &arp); err != nil {
		return nil, err
	}
	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > time.Second*3 {
			return nil, errors.New("timeout getting ARP reply")
		}
		data, _, err := s.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return nil, err
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			if net.IP(arp.SourceProtAddress).Equal(net.IP(arpDst)) {
				return net.HardwareAddr(arp.SourceHwAddress), nil
			}
		}
	}
}

// scan scans the dst IP address of this scanner.
func (s *client) scan(dstPort, srcPort layers.TCPPort) error {
	// First off, get the MAC address we should be sending packets to.
	hwaddr, err := s.getHwAddr()
	if err != nil {
		return err
	}
	// Construct all the network layers we need.
	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       hwaddr,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := layers.IPv4{
		SrcIP:    s.fakeSrc,
		DstIP:    s.dst,
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}
	tcp := layers.TCP{
		SrcPort: srcPort,
		DstPort: dstPort, // will be incremented during the scan
		SYN:     true,
	}
	err = tcp.SetNetworkLayerForChecksum(&ip4)
	if err != nil {
		panic(err)
	}

	if err := s.send(&eth, &ip4, &tcp); err != nil {
		return fmt.Errorf("error sending to port %v: %v", tcp.DstPort, err)
	}
	return nil
}

// send sends the given layers as a single packet on the network.
func (s *client) send(l ...gopacket.SerializableLayer) error {
	if err := gopacket.SerializeLayers(s.buf, s.opts, l...); err != nil {
		return err
	}
	return s.handle.WritePacketData(s.buf.Bytes())
}

func main() {
	defer util.Run()()
	router, err := routing.New()
	if err != nil {
		log.Fatal("routing error:", err)
	}

	dstIPStr, srcIPStr := "154.208.143.31", "123.4.5.6"
	dsrPort, srcPort := 1234, 9999
	dstIP, err := tools.String2IPV4(dstIPStr)
	if err != nil {
		panic(err)
	}
	srcIP, err := tools.String2IPV4(srcIPStr)
	if err != nil {
		panic(err)
	}

	// Note:  newScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	s, err := NewClient(dstIP, srcIP, router)
	if err != nil {
		err = fmt.Errorf("unable to create client for %v: %v", dstIP, err)
		panic(err)
	}
	defer s.Close()
	if err := s.scan(tools.Int2TCPPort(dsrPort), tools.Int2TCPPort(srcPort)); err != nil {
		log.Printf("unable to scan %v: %v", dstIP, err)
	}
}
