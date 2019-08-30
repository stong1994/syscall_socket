package client

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

type UDTHeader struct {
	SrcPort  uint16
	DstPort  uint16
	Length   uint16
	Checksum uint16
}

type IPHeader struct {
	VersionIHL uint8 // 版本 + 首部长度 4+4 bite 版本：目前的IP协议版本号为4，下一代IP协议版本号为6。 首部长度: 数据包头部长度。它表示数据包头部包括多少个32位长整型，也就是多少个4字节的数据
	//IHL        uint8 //
	TOS    uint8  // 服务类型 8 bite
	Length uint16 // 总长度 16 bite 首部长度+数据长度
	Id     uint16 // 标识 16 bite
	// todo 先让flag和offset共用
	// Flags      uint8 // 标志 3 bite 共3位。R、DF、MF三位。目前只有后两位有效，DF位：为1表示不分片，为0表示分片。MF：为1表示“更多的片”，为0表示这是最后一片。
	FragOffset uint16 // 片位移 13 bite ：本分片在原先数据报文中相对首位的偏移位。（需要再乘以8）
	TTL        uint8  // 生存时间 8 bite
	Protocol   uint8  // 协议 8 bite TCP的协议号为6，UDP的协议号为17。ICMP的协议号为1，IGMP的协议号为2
	Checksum   uint16 // 16 bite
	SrcIP      uint32 // 32 bite
	DstIP      uint32 // 32 bite
}

type EthernetHeader struct {
	SrcMAC       []byte
	DstMAC       []byte
	EthernetType uint16
}
