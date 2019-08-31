package client

import (
	"bytes"
	"encoding/binary"
	"github.com/google/gopacket/layers"
	tools "tcp_server/toos"
	"unsafe"
)

const (
	FIN = 1  // 00 0001
	SYN = 2  // 00 0010
	RST = 4  // 00 0100
	PSH = 8  // 00 1000
	ACK = 16 // 01 0000
	URG = 32 // 10 0000
)

type ITransferLayer interface {
	SerializeHeader(ip *IPHeader) ([]byte, error)
}

type EthernetHeader struct {
	SrcMAC       []byte
	DstMAC       []byte
	EthernetType uint16
}

func NewEthernetHeader(srcMac, dstMac []byte, ethernetType uint16) *EthernetHeader {
	return &EthernetHeader{
		SrcMAC:       srcMac,
		DstMAC:       dstMac,
		EthernetType: ethernetType,
	}
}

func SerializeEthernetHeader(eth *EthernetHeader) ([]byte, error) {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, eth.DstMAC)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, eth.SrcMAC)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, eth.EthernetType)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// TODO 不支持可选部分
type IPHeader struct {
	Version    uint8  // 版本 4 bite 版本：目前的IP协议版本号为4，下一代IP协议版本号为6。
	IHL        uint8  // 首部长度 4 bite: 数据包头部长度。它表示数据包头部包括多少个32位长整型，也就是多少个4字节的数据
	TOS        uint8  // 服务类型 8 bite
	Length     uint16 // 总长度 16 bite 首部长度+数据长度
	Id         uint16 // 标识 16 bite
	Flags      uint8  // 标志 3 bite 共3位。R、DF、MF三位。目前只有后两位有效，DF位：为1表示不分片，为0表示分片。MF：为1表示“更多的片”，为0表示这是最后一片。
	FragOffset uint16 // 片位移 13 bite ：本分片在原先数据报文中相对首位的偏移位。（需要再乘以8）
	TTL        uint8  // 生存时间 8 bite
	Protocol   uint8  // 协议 8 bite TCP的协议号为6，UDP的协议号为17。ICMP的协议号为1，IGMP的协议号为2
	Checksum   uint16 // 16 bite
	SrcIP      uint32 // 32 bite
	DstIP      uint32 // 32 bite
}

func NewIPHeader(srcHost uint32, dstHost uint32, protocol uint8, dataLen uint16) *IPHeader {
	return &IPHeader{
		Version:  4,
		IHL:      5,
		Length:   20 + dataLen,
		TTL:      64,
		Protocol: protocol,
		SrcIP:    srcHost,
		DstIP:    dstHost,
	}
}

func SerializeIPHeader(ip *IPHeader) []byte {
	res := make([]byte, 20)
	res[0] = ip.Version<<4 + ip.IHL
	res[1] = ip.TOS
	res[2] = uint8(ip.Length >> 8)
	res[3] = uint8(ip.Length)
	res[4] = uint8(ip.Id >> 8)
	res[5] = uint8(ip.Id)
	flagAndOffset := uint16(uint16(ip.Flags)<<13) + ip.FragOffset
	res[6] = uint8(flagAndOffset >> 8)
	res[7] = uint8(flagAndOffset)
	res[8] = ip.TTL
	res[9] = ip.Protocol
	res[12:] = tools.IntToBytes(ip.SrcIP)
	res[16:] = tools.IntToBytes(ip.DstIP)
	ip.Checksum = checkSum(res)
	res[10] = uint8(ip.Checksum >> 8)
	res[11] = uint8(ip.Checksum)
	return res
}

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

func NewTcpHeader(srcPort uint16, dstPort uint16, seqNum, ackNum uint32, flag uint8, windows, urgentPtr uint16) *TCPHeader {
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
		UrgentPtr: urgentPtr,
	}
}

func (tcp *TCPHeader) SerializeHeader(ip *IPHeader) ([]byte, error) {
	var buffer bytes.Buffer
	psdHeader := NewPsdHeader(ip.SrcIP, ip.DstIP, uint8(layers.IPProtocolTCP))
	err := binary.Write(&buffer, binary.BigEndian, psdHeader)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, tcp)
	if err != nil {
		return nil, err
	}
	sum := checkSum(buffer.Bytes())
	tcp.Checksum = sum

	buffer.Reset()
	err = binary.Write(&buffer, binary.BigEndian, tcp)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type UDPHeader struct {
	SrcPort  uint16
	DstPort  uint16
	Length   uint16
	Checksum uint16
}

func NewUDPHeader(srcPort, dstPort, dataLen uint16) *UDPHeader {
	return &UDPHeader{
		SrcPort: srcPort,
		DstPort: dstPort,
		Length:  8 + dataLen,
	}
}

func (udp *UDPHeader) SerializeHeader(ip *IPHeader) ([]byte, error) {
	var buffer bytes.Buffer
	psdHeader := NewPsdHeader(ip.SrcIP, ip.DstIP, uint8(layers.IPProtocolUDP))
	err := binary.Write(&buffer, binary.BigEndian, psdHeader)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buffer, binary.BigEndian, udp)
	if err != nil {
		return nil, err
	}
	sum := checkSum(buffer.Bytes())
	udp.Checksum = sum

	buffer.Reset()
	err = binary.Write(&buffer, binary.BigEndian, udp)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// TCP和UDP需要一个伪头部来计算校验和
type PsdHeader struct {
	SrcIP     uint32
	DstIP     uint32
	Zero      uint8
	ProtoType uint8
	Length    uint16
}

func NewPsdHeader(srcIP, dstIP uint32, protocol uint8) *PsdHeader {
	return &PsdHeader{
		SrcIP:     srcIP,
		DstIP:     dstIP,
		Zero:      0,
		ProtoType: protocol,
		Length:    0,
	}
}
