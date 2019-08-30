package main

import (
	"fmt"
	"github.com/google/gopacket/layers"
	"net"
	"tcp_server/api"
	"tcp_server/client"
	tools "tcp_server/toos"
)

func main() {
	var c api.IClient
	srcIP, dstIP, srcPort, dstPort := param()
	c, err := client.NewFakeClient(srcIP, dstIP, srcPort, dstPort)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	tcp := layers.TCP{
		//BaseLayer: layers.BaseLayer{IntToBytes(100), IntToBytes(32)},
		SrcPort:    srcPort,
		DstPort:    dstPort,
		Seq:        102,
		Ack:        103,
		SYN:        true,
		Window:     6666,
		//Checksum:   7777,
		Urgent:     1,
	}
	udp := createUDPLayer()
	_ = udp
	send, err := c.Send(&tcp)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("len", send)
}

func param() (srcIP, dstIP net.IP, srcPort, dstPort layers.TCPPort) {
	srcIP, _ = tools.String2IPV4("1,2,3,4")
	dstIP, _ = tools.String2IPV4("154.208.143.31")
	srcPort = tools.Int2TCPPort(123)
	dstPort = tools.Int2TCPPort(9999)
	return
}

func createUDPLayer() (udp *layers.UDP) {
	udp = &layers.UDP{}
	udp.SrcPort = layers.UDPPort(12345)
	udp.DstPort = layers.UDPPort(9999)
	return
}