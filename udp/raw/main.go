package main

import (
	"golang.org/x/net/ipv4"
	"log"
	"net"
)

func checkSum(msg []byte) uint16 {
	sum := 0
	for n := 1; n < len(msg)-1; n += 2 {
		sum += int(msg[n])*256 + int(msg[n+1])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var ans = uint16(^sum)
	return ans
}

func ipHeader(dst, src net.IP, buff []byte) *ipv4.Header {
	//填充ip首部
	iph := &ipv4.Header{
		Version:  ipv4.Version,
		//IP头长一般是20
		Len:      ipv4.HeaderLen,
		TOS:      0x00,
		//buff为数据
		TotalLen: ipv4.HeaderLen + len(buff),
		TTL:      64,
		Flags:    ipv4.DontFragment,
		FragOff:  0,
		Protocol: 17,
		Checksum: 0,
		Src:      src,
		Dst:      dst,
	}

	h, err := iph.Marshal()
	if err != nil {
		log.Fatalln(err)
	}
	//计算IP头部校验值
	iph.Checksum = int(checkSum(h))
	return iph
}

func udpHeader(dst, src net.IP, buff []byte) []byte {
	//填充udp首部
	//udp伪首部
	udph := make([]byte, 20)
	//源ip地址
	udph[0], udph[1], udph[2], udph[3] = src[12], src[13], src[14], src[15]
	//目的ip地址
	udph[4], udph[5], udph[6], udph[7] = dst[12], dst[13], dst[14], dst[15]
	//协议类型
	udph[8], udph[9] = 0x00, 0x11
	//udp头长度
	udph[10], udph[11] = 0x00, byte(len(buff)+8)
	//下面开始就真正的udp头部
	//源端口号
	udph[12], udph[13] = 0x27, 0x10
	//目的端口号
	udph[14], udph[15] = 0x17, 0x70
	//udp头长度
	udph[16], udph[17] = 0x00, byte(len(buff)+8)
	//校验和
	udph[18], udph[19] = 0x00, 0x00
	//计算校验值
	check := checkSum(append(udph, buff...))
	udph[18], udph[19] = byte(check>>8&255), byte(check&255)
	return udph
}

func send()  {
	listener, err := net.ListenPacket("ip4:udp", "192.168.1.104")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	//listener 实现了net.PacketConn接口
	r, err := ipv4.NewRawConn(listener)
	if err != nil {
		log.Fatal(err)
	}

	// 数据
	data := []byte(string("data"))
	//目的IP
	dst := net.IPv4(154, 208, 143, 31)
	//源IP
	src := net.IPv4(128,208, 234, 123)

	iph := ipHeader(dst, src, data)
	udph := udpHeader(dst, src, data)
	//发送自己构造的UDP包
	if err = r.WriteTo(iph, append(udph[12:20], data...), nil); err != nil {
		log.Fatal(err)
	}
}

func main()  {
	send()
}