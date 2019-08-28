package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net"
	"syscall"
)

type NameTag struct {
	Name string `json:"name"`
}

func main() {
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
	err = syscall.Sendto(fd, p, 0, &addr)
	if err != nil {
		log.Fatal("Sendto:", err)
	}
}

func pkt() []byte {
	sb := gopacket.NewSerializeBuffer()
	nt := NameTag{"Paul"}
	b, _ := json.Marshal(nt)
	pld := gopacket.Payload(b)
	l := uint16(len(pld))
	udp := layers.UDP{
		SrcPort:  8888,
		DstPort:  9999,
		Length:   l + 8,
		Checksum: 0,
	}
	l = l + 8
	ip := layers.IPv4{
		Version:    0x4,
		IHL:        5,
		Length:     20 + l,
		TTL:        255,
		Flags:      0x40,
		FragOffset: 0,
		Checksum:   0,
		Protocol:   syscall.IPPROTO_UDP,
		DstIP:      net.IPv4(154, 208, 143, 31),
		SrcIP:      net.IPv4(127, 0, 0, 1),
	}
	l = l + 20

	//eth := layers.Ethernet{
	//	EthernetType: layers.EthernetTypeIPv4,
	//	SrcMAC: net.HardwareAddr{
	//		0X--, 0x--, 0x--, 0x--, 0x--, 0x--,
	//	},
	//	DstMAC: net.HardwareAddr{
	//		0x--, 0x--, 0x--, 0x--, 0x--, 0x--,
	//	},
	//}
	fmt.Println(pld.SerializeTo(sb, gopacket.SerializeOptions{}))
	fmt.Println(udp.SerializeTo(sb, gopacket.SerializeOptions{}))
	fmt.Println(ip.SerializeTo(sb, gopacket.SerializeOptions{}))
	//fmt.Println(eth.SerializeTo(sb, gopacket.SerializeOptions{}))

	//Debug prints here (first line is for dump to packet validator)
	fmt.Println(hex.EncodeToString(sb.Bytes()))
	fmt.Println(sb.Bytes())

	return sb.Bytes()
}
