package client

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/routing"
	"net"
	tools "tcp_server/toos"
)

type rawFakeClient struct {
	srcIP   net.IP
	srcPort layers.TCPPort
	dstPort layers.TCPPort
	*client
}

func NewRawFakeClient(srcIP, dstIP net.IP, srcPort, dstPort layers.TCPPort) (*rawFakeClient, error) {
	defer util.Run()()
	router, err := routing.New()
	if err != nil {
		return nil, err
	}
	client, err := newClient(dstIP, router)
	return &rawFakeClient{
		srcIP:   srcIP,
		srcPort: srcPort,
		dstPort: dstPort,
		client:  client,
	}, nil
}

func (fc *rawFakeClient) Close() {
	fc.client.close()
}

// transferLayer 为TCP或者UDP Layer
func (r *rawFakeClient) Send(transferLayer, payloadLayer gopacket.SerializableLayer) (int, error) {
	srcIP, err := tools.BytesToInt(r.srcIP)
	if err != nil {
		return 0, err
	}
	dstIP, err := tools.BytesToInt(r.dstIP)
	if err != nil {
		return 0, err
	}
	payload := payloadLayer.(*gopacket.Payload).Payload()

	resolveLayer, protocol, err := resolveTransferHeader(transferLayer, uint16(len(payload)))
	if err != nil {
		return 0, err
	}

	ipHeader := NewIPHeader(srcIP, dstIP, uint8(protocol), uint16(len(payload)))

	dstMac, err := r.getHwAddr()
	if err != nil {
		return 0, err
	}
	ethernetHeader := NewEthernetHeader(r.iface.HardwareAddr, dstMac, uint16(layers.EthernetTypeIPv4))

	data, err := newPackageData(ethernetHeader, ipHeader, resolveLayer, payload)
	if err != nil {
		return 0, err
	}

	err = r.send(data)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func resolveTransferHeader(transferLayer gopacket.SerializableLayer, dataLen uint16) (ITransferLayer, layers.IPProtocol, error) {
	var (
		tcpHeader *TCPHeader
		udpHeader *UDPHeader
	)
	switch transferLayer.(type) {
	case *layers.TCP:
		tcp := transferLayer.(*layers.TCP)
		tcpHeader = NewTcpHeader(uint16(tcp.SrcPort), uint16(tcp.DstPort), uint32(tcp.Seq), uint32(tcp.Ack), calcFlag(tcp), uint16(tcp.Window), uint16(tcp.Urgent))
		return tcpHeader, layers.IPProtocolTCP, nil
	case *layers.UDP:
		udp := transferLayer.(*layers.UDP)
		udpHeader = NewUDPHeader(uint16(udp.SrcPort), uint16(udp.DstPort), dataLen)
		return udpHeader, layers.IPProtocolUDP, nil
	}
	return nil, 0, fmt.Errorf("not valid type of transfer layer")
}

func calcFlag(tcp *layers.TCP) uint8 {
	var flag uint8
	if tcp.FIN {
		flag += FIN
	}
	if tcp.SYN {
		flag += SYN
	}
	if tcp.RST {
		flag += RST
	}
	if tcp.PSH {
		flag += PSH
	}
	if tcp.ACK {
		flag += ACK
	}
	if tcp.URG {
		flag += URG
	}
	return flag
}

func checkSum(data []byte) uint16 {
	var (
		sum    uint32
		length = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += sum >> 16
	return uint16(^sum)
}

func newPackageData(ethHeader *EthernetHeader, ipHeader *IPHeader, transferHead ITransferLayer, data []byte) ([]byte, error) {
	var (
		eth, ip, transferByte []byte
		err                   error
	)

	eth, err = SerializeEthernetHeader(ethHeader)
	if err != nil {
		return nil, err
	}
	ip = SerializeIPHeader(ipHeader)

	switch transferHead.(type) {
	case *TCPHeader:
		transferByte, err = transferHead.SerializeHeader(ipHeader)
	case *UDPHeader:
		transferByte, err = transferHead.SerializeHeader(ipHeader)
	default:
		err = fmt.Errorf("not valid type of transfer header")
	}
	if err != nil {
		return nil, err
	}

	res := eth
	res = append(res, ip...)
	res = append(res, transferByte...)
	return res, nil
}
