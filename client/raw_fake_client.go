package client

import (
	"bytes"
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/routing"
	"net"
	tools "tcp_server/toos"
	"unsafe"
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

// transferLayer 为TCP或者UDP
func (r *rawFakeClient) Send(transferLayer, payloadLayer gopacket.SerializableLayer) (int, error) {
	srcIP, err := tools.BytesToInt(r.srcIP)
	if err != nil {
		return 0, err
	}
	dstIP, err := tools.BytesToInt(r.dstIP)
	if err != nil {
		return 0, err
	}

	ipHeader := newIPHeader(srcIP, dstIP)

	payload := payloadLayer.(*gopacket.Payload).Payload()

	ethernetHeader, err := r.newEthernetHeader()
	if err != nil {
		return 0, err
	}
	data, err := newTcpData(ethernetHeader, ipHeader, transferHeader(transferLayer), payload)
	if err != nil {
		return 0, err
	}
	err = r.send(data)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func transferHeader(transferLayer gopacket.SerializableLayer) *TCPHeader {
	var (
		tcpHeader = TCPHeader{
			Offset: uint8(uint16(unsafe.Sizeof(TCPHeader{}))/4) << 4,
		}
	)
	switch transferLayer.(type) {
	case *layers.TCP:
		tcp := transferLayer.(*layers.TCP)
		tcpHeader.SrcPort = uint16(tcp.SrcPort)
		tcpHeader.DstPort = uint16(tcp.DstPort)
		tcpHeader.Window = uint16(tcp.Window)
		tcpHeader.AckNum = uint32(tcp.Ack)
		tcpHeader.SeqNum = uint32(tcp.Seq)
		tcpHeader.UrgentPtr = uint16(tcp.Urgent)
		tcpHeader.Flag = calcFlag(tcp)
	}
	return &tcpHeader
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

func converEthernetHeader(eth *EthernetHeader) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, eth.DstMAC)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&buffer, binary.BigEndian, eth.SrcMAC)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&buffer, binary.BigEndian, eth.EthernetType)
	if err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

func (r *rawFakeClient) newEthernetHeader() (*EthernetHeader, error) {
	hwaddr, err := r.getHwAddr()
	if err != nil {
		return nil, err
	}
	return &EthernetHeader{
		SrcMAC:       r.iface.HardwareAddr,
		DstMAC:       hwaddr,
		EthernetType: uint16(layers.EthernetTypeIPv4),
	}, nil
}

func newIPHeader(srcHost uint32, dstHost uint32) *IPHeader {
	return &IPHeader{
		SrcIP:      srcHost,
		DstIP:      dstHost,
		VersionIHL: uint8(4)<<4 + 5,
		Protocol:   6,
		TTL:        64,
	}
}

func newTcpHeader(srcPort uint16, dstPort uint16, seqNum, ackNum uint32, flag uint8, windows uint16) *TCPHeader {
	// 创建TCP内部包头内容
	return &TCPHeader{
		SrcPort: srcPort,
		DstPort: dstPort,
		SeqNum:  seqNum,
		AckNum:  ackNum,
		// offset(4bite) + 保留字段(6bite)+flag(6bite) = 16bite
		// offset只占四bite，后四个bite为保留字段所有，因此要向左移4位
		// flag实际上只占6bite，前两个bite为保留字段所有
		Offset: uint8(uint16(unsafe.Sizeof(TCPHeader{}))/4) << 4,
		Flag:   flag,
		Window: windows,
	}
}

func newTcpData(ethHeader *EthernetHeader, psdHeader *IPHeader, tcpHeader *TCPHeader, data []byte) ([]byte, error) {
	eth := converEthernetHeader(ethHeader)
	var buffer bytes.Buffer
	// buffer用来写入两种首部来求得校验和
	err := binary.Write(&buffer, binary.BigEndian, eth)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, psdHeader)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, tcpHeader)
	if err != nil {
		return nil, err
	}

	tcpHeader.Checksum = checkSum(buffer.Bytes())
	// 接下来清空buffer，填充实际要发送的部分
	buffer.Reset()
	err = binary.Write(&buffer, binary.BigEndian, eth)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, psdHeader)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, tcpHeader)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), err
}
