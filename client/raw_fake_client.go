package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/routing"
	"net"
	"strconv"
	"strings"
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

// layer 为TCP或者UDP
func (r *rawFakeClient) Send(transferLayer, payloadLayer gopacket.SerializableLayer) (int, error) {
	srcIP, err := tools.BytesToInt(r.srcIP)
	if err != nil {
		return 0, err
	}
	dstIP, err := tools.BytesToInt(r.dstIP)
	if err != nil {
		return 0, err
	}

	srcIP = inetAddr("1.2.3.4")
	dstIP = inetAddr("154.208.143.31")

	psdHeader := newPsdHeader(srcIP, dstIP)

	payload := payloadLayer.(*gopacket.Payload).Payload()

	ethernetHeader, err := r.newEthernetHeader()
	if err != nil {
		return 0, err
	}
	data, err := newTcpData(ethernetHeader, psdHeader, transferHeader(transferLayer), payload)
	err = r.send(data)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func transferHeader(transferLayer gopacket.SerializableLayer) (*TCPHeader) {
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

const (
	FIN = 1  // 00 0001
	SYN = 2  // 00 0010
	RST = 4  // 00 0100
	PSH = 8  // 00 1000
	ACK = 16 // 01 0000
	URG = 32 // 10 0000
)

type TCPHeader struct {
	SrcPort   uint16
	DstPort   uint16
	SeqNum    uint32
	AckNum    uint32
	Offset    uint8
	Flag      uint8
	Window    uint16
	Checksum  uint16
	UrgentPtr uint16
}

type PsdHeader struct {
	SrcAddr   uint32
	DstAddr   uint32
	Zero      uint8
	ProtoType uint8
	TcpLength uint16
	Version   uint8
	TTL       uint8
}

type EthernetHeader struct {
	SrcMAC       []byte
	DstMAC       []byte
	EthernetType uint16
}

func inetAddr(host string) uint32 {
	var (
		segments = strings.Split(host, ".")
		ip       [4]uint64
		ret      uint64
	)
	if len(segments) != 4 {
		fmt.Println(host)
		return 0
	}
	for i := 0; i < 4; i++ {
		ip[i], _ = strconv.ParseUint(segments[i], 10, 64)
	}
	ret = ip[3]<<24 + ip[2]<<16 + ip[1]<<8 + ip[0]
	return uint32(ret)
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

func (r *rawFakeClient) newEthernetHeader() (*EthernetHeader, error) {
	hwaddr, err := r.getHwAddr()
	if err != nil {
		return nil, err
	}
	return &EthernetHeader{
		SrcMAC: r.iface.HardwareAddr,
		DstMAC: hwaddr,
		EthernetType: uint16(layers.EthernetTypeIPv4),
	}, nil
}

func newPsdHeader(srcHost uint32, dstHost uint32) *PsdHeader {
	return &PsdHeader{
		SrcAddr: srcHost,
		DstAddr: dstHost,
		Version: 4,
		TTL:     64,
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

func newTcpData(ethHeader *EthernetHeader, psdHeader *PsdHeader, tcpHeader *TCPHeader, data []byte) ([]byte, error) {
	psdHeader.TcpLength = uint16(unsafe.Sizeof(TCPHeader{})) + uint16(len(data))
	var buffer bytes.Buffer
	/*buffer用来写入两种首部来求得校验和*/
	err := binary.Write(&buffer, binary.BigEndian, psdHeader)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, tcpHeader)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, ethHeader)
	if err != nil {
		return nil, err
	}
	tcpHeader.Checksum = checkSum(buffer.Bytes())
	/*接下来清空buffer，填充实际要发送的部分*/
	buffer.Reset()
	err = binary.Write(&buffer, binary.BigEndian, ethHeader)
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
