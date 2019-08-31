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

// 实际上做了两件事：构造ipv4 和 给layer设置计算checksum的网络层layer （违反单一责任原则）
func (g *gopacketFakeClient) constructIPv4(layer gopacket.SerializableLayer) (*layers.IPv4, error) {
	var (
		err error
		ip4 = layers.IPv4{
			SrcIP:   g.srcIP,
			DstIP:   g.dstIP,
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
func (g *gopacketFakeClient) Send(transferLayer, payloadLayer gopacket.SerializableLayer) (int, error) {
	var (
		data []byte
		err  error
	)
	// First off, get the MAC address we should be sending packets to.
	hwaddr, err := g.client.getHwAddr()
	if err != nil {
		return 0, err
	}
	// Construct all the network layers we need.
	eth := layers.Ethernet{
		SrcMAC:       g.iface.HardwareAddr,
		DstMAC:       hwaddr,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4, err := g.constructIPv4(transferLayer)
	if err != nil {
		return 0, err
	}
	if data, err = g.serializeData(&eth, ip4, transferLayer, payloadLayer); err != nil {
		return 0, fmt.Errorf("error sending to port %v: %v", g.dstPort, err)
	}
	err = g.send(data)
	if err != nil {
		return 0, err
	}
	return len(g.buf.Bytes()), nil
}
