package test

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"syscall_socket/client"
	tools "syscall_socket/toos"
	"testing"
)

func TestTCPPackage(t *testing.T) {
	tcp := layers.TCP{
		SrcPort: 22,
		DstPort: 21,
		Seq:     102,
		Ack:     103,
		SYN:     true,
		Window:  6666,
		Urgent:  1,
	}
	var (
		data  = "hello is me"
		srcIP = "1.2.3.4"
		dstIP = "4.3.2.1"
	)
	sIP, _ := tools.String2IPV4(srcIP)
	dIP, _ := tools.String2IPV4(dstIP)
	c, err := client.NewGopacketFakeClient(sIP, dIP, tcp.SrcPort, tcp.DstPort)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	payload := gopacket.Payload([]byte(data))
	_, err = c.Send(&tcp, &payload)
	if err != nil {
		t.Error(err)
	}
}

func TestUDPPackage(t *testing.T) {
	udp := layers.UDP{
		SrcPort: 22,
		DstPort: 21,
	}
	var (
		data  = "hello is me"
		srcIP = "1.2.3.4"
		dstIP = "4.3.2.1"
	)
	sIP, _ := tools.String2IPV4(srcIP)
	dIP, _ := tools.String2IPV4(dstIP)
	c, err := client.NewGopacketFakeClient(sIP, dIP, layers.TCPPort(udp.SrcPort), layers.TCPPort(udp.DstPort))
	if err != nil {
		panic(err)
	}
	defer c.Close()

	payload := gopacket.Payload([]byte(data))
	_, err = c.Send(&udp, &payload)
	if err != nil {
		t.Error(err)
	}
}
