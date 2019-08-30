package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"syscall"
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

func CheckSum(data []byte) uint16 {
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

func NewPsdHeader(srcHost string, dstHost string) *PsdHeader {
	// 创建TCP外部包头内容
	return &PsdHeader{
		SrcAddr:   inetAddr(srcHost),
		DstAddr:   inetAddr(dstHost),
		Zero:      0,
		ProtoType: syscall.IPPROTO_TCP,
		TcpLength: 0,
	}
}

func NewTcpHeader(srcPort uint16, dstPort uint16) *TCPHeader {
	// 创建TCP内部包头内容
	return &TCPHeader{
		SrcPort: srcPort,
		DstPort: dstPort,
		SeqNum:  102,
		AckNum:  101,
		// offset(4bite) + 保留字段(6bite)+flag(6bite) = 16bite
		// offset只占四bite，后四个bite为保留字段所有，因此要向左移4位
		// flag实际上只占6bite，前两个bite为保留字段所有
		Offset:   uint8(uint16(unsafe.Sizeof(TCPHeader{}))/4) << 4,
		Flag:     SYN,
		Window:   60000,
		Checksum: 0,
	}
}

func NewTcpData(psdHeader *PsdHeader, tcpHeader *TCPHeader, data string) []byte {
	psdHeader.TcpLength = uint16(unsafe.Sizeof(TCPHeader{})) + uint16(len(data))
	var buffer bytes.Buffer
	/*buffer用来写入两种首部来求得校验和*/
	err := binary.Write(&buffer, binary.BigEndian, psdHeader)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&buffer, binary.BigEndian, tcpHeader)
	if err != nil {
		panic(err)
	}
	tcpHeader.Checksum = CheckSum(buffer.Bytes())
	/*接下来清空buffer，填充实际要发送的部分*/
	buffer.Reset()
	err = binary.Write(&buffer, binary.BigEndian, tcpHeader)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&buffer, binary.BigEndian, []byte(data))
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

func Send(srcIp string, srcPort uint16, dstIp string, dstPort uint16, data string) (exec func() error, close func(), err error) {
	sockfd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, nil, err
	}
	var addr syscall.SockaddrInet4
	// 解析ip
	for i, s := range strings.SplitN(dstIp, ".", 4) {
		if d, err := strconv.ParseInt(s, 10, 64); err != nil {
			return nil, nil, err
		} else {
			addr.Addr[i] = byte(d)
		}
	}
	addr.Port = int(dstPort)

	return func() error {
			if err = syscall.Sendto(sockfd, NewTcpData(NewPsdHeader(srcIp, dstIp), NewTcpHeader(srcPort, dstPort), data), 0, &addr); err != nil {
				return err
			}
			return nil
		}, func() {
			syscall.Shutdown(sockfd, syscall.SHUT_RDWR)
		}, nil
}

func main() {
	if exec, close, err := Send("128.208.234.123", 234, "154.208.143.31", 9999, "ac fun"); err != nil {
		fmt.Println("err is :", err)
	} else {
		defer close()
		err := exec()
		if err != nil {
			fmt.Println("exec:", err)
		}
	}
	fmt.Println("success")
}
