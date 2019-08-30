package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net"
	"syscall"
)

func main() {
	//srcIP := net.ParseIP("1.2.3.4")
	//dstIP := net.ParseIP("")
	var err error
	fd, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_UDP)
	if e != nil {
		fmt.Println("Problem @ location 1")
	}
	addr := syscall.SockaddrInet4{
		Port: 9999,
		Addr: [4]byte{154, 208, 143, 31},
	}
	p := pkt()
	//fmt.Println("result\n", p.String())
	err = syscall.Sendto(fd, p, 0, &addr)
	if err != nil {
		log.Fatal("Sendto:", err)
	}
}

func createIPv4ChecksumTestLayer() (ip4 *layers.IPv4) {
	ip4 = &layers.IPv4{}
	ip4.Version = 4
	ip4.TTL = 64
	ip4.SrcIP = net.ParseIP("192.0.2.1")
	ip4.DstIP = net.ParseIP("154.208.143.31")
	return
}

func createUDPChecksumTestLayer() (udp *layers.UDP) {
	udp = &layers.UDP{}
	udp.SrcPort = layers.UDPPort(12345)
	udp.DstPort = layers.UDPPort(9999)
	return
}

func pkt() []byte {
	ip4 := createIPv4ChecksumTestLayer()
	ip4.Protocol = layers.IPProtocolUDP

	udp := createUDPChecksumTestLayer()
	err := udp.SetNetworkLayerForChecksum(ip4)
	if err != nil {
		panic(err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: false}

	//payload := gopacket.Payload([]byte("meowmeowmeow"))
	err = gopacket.SerializeLayers(buf, opts, udp)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}
