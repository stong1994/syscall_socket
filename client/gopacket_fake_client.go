package client

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/routing"
	"net"
)

type gopacketFakeClient struct {
	srcIP   net.IP
	srcPort layers.TCPPort
	dstPort layers.TCPPort
	*client
}

func NewGopacketFakeClient(srcIP, dstIP net.IP, srcPort, dstPort layers.TCPPort) (*gopacketFakeClient, error) {
	defer util.Run()()
	router, err := routing.New()
	if err != nil {
		return nil, err
	}
	client, err := newClient(dstIP, router)
	return &gopacketFakeClient{
		srcIP:   srcIP,
		srcPort: srcPort,
		dstPort: dstPort,
		client:  client,
	}, nil
}

func (fc *gopacketFakeClient) Close() {
	fc.client.close()
}

// 实际上做了两件事：构造ipv4 和 给layer计算checksum （违反单一责任原则）
func (s *gopacketFakeClient) constructIPv4(layer gopacket.SerializableLayer) (*layers.IPv4, error) {
	var (
		err error
		ip4 = layers.IPv4{
			SrcIP:   s.srcIP,
			DstIP:   s.dstIP,
			Version: 4,
			TTL:     64,
		}
	)
	// check layer type
	switch layer.(type) {
	case *layers.TCP:
		ip4.Protocol = layers.IPProtocolTCP
		err = layer.(*layers.TCP).SetNetworkLayerForChecksum(&ip4)
		if err != nil {
			return nil, err
		}
	case *layers.UDP:
		ip4.Protocol = layers.IPProtocolUDP
		err = layer.(*layers.UDP).SetNetworkLayerForChecksum(&ip4)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("layer must be tcp or udp")
	}
	return &ip4, nil
}

// transferLayer 为TCP或者UDP
func (s *gopacketFakeClient) Send(transferLayer, payloadLayer gopacket.SerializableLayer) (int, error) {
	var (
		data []byte
		err  error
	)
	// First off, get the MAC address we should be sending packets to.
	hwaddr, err := s.client.getHwAddr()
	if err != nil {
		return 0, err
	}
	// Construct all the network layers we need.
	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       hwaddr,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4, err := s.constructIPv4(transferLayer) // todo layer是一个接口作为参数传递是否会影响外部的值？应该会
	if err != nil {
		return 0, err
	}
	if data, err = s.serializeData(&eth, ip4, transferLayer, payloadLayer); err != nil {
		return 0, fmt.Errorf("error sending to port %v: %v", s.dstPort, err)
	}
	err = s.send(data)
	if err != nil {
		return 0, err
	}
	return len(s.buf.Bytes()), nil
}

//func main() {
//	defer util.Run()()
//	router, err := routing.New()
//	if err != nil {
//		log.Fatal("routing error:", err)
//	}
//
//	dstIPStr, srcIPStr := "154.208.143.31", "123.4.5.6"
//	dsrPort, srcPort := 1234, 9999
//	dstIP, err := tools.String2IPV4(dstIPStr)
//	if err != nil {
//		panic(err)
//	}
//	srcIP, err := tools.String2IPV4(srcIPStr)
//	if err != nil {
//		panic(err)
//	}
//
//	// Note:  newScanner creates and closes a pcap Handle once for
//	// every scan target.  We could do much better, were this not an
//	// example ;)
//	s, err := NewClient(dstIP, srcIP, router)
//	if err != nil {
//		err = fmt.Errorf("unable to create client for %v: %v", dstIP, err)
//		panic(err)
//	}
//	defer s.Close()
//	if err := s.scan(tools.Int2TCPPort(dsrPort), tools.Int2TCPPort(srcPort)); err != nil {
//		log.Printf("unable to scan %v: %v", dstIP, err)
//	}
//}
