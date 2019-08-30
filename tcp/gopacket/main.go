package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net"
	"syscall"
)

func main() {
	var err error
	fd, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
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

func pkt() []byte {
	ip4 := createIPv4ChecksumTestLayer()
	ip4.Protocol = layers.IPProtocolTCP

	tcp := layers.TCP{
		//BaseLayer: layers.BaseLayer{IntToBytes(100), IntToBytes(32)},
		SrcPort:    8888,
		DstPort:    9999,
		Seq:        102,
		Ack:        103,
		SYN:        true,
		Window:     6666,
		//Checksum:   7777,
		Urgent:     1,
	}
	err := tcp.SetNetworkLayerForChecksum(ip4)
	if err != nil {
		panic(err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: false}

	payload := gopacket.Payload("hello is me")
	fmt.Println(payload.Payload())
	err = gopacket.SerializeLayers(buf, opts, &tcp, payload)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

//整形转换成字节
func IntToBytes(n int32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.BigEndian, n)
	if err != nil {
		panic(err)
	}
	return bytesBuffer.Bytes()
}
