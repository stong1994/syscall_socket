package main

import (
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"tcp_server/api"
	"tcp_server/client"
	tools "tcp_server/toos"
)

var (
	srcIP, dstIP     string
	srcPort, dstPort int
)

func init() {
	flag.Parse()
	flag.StringVar(&srcIP, "srcIP", "4.3.2.1", "源IP")
	flag.StringVar(&dstIP, "dstIP", "1.2.3.4", "目标IP")
	flag.IntVar(&srcPort, "srcPort", 22, "源端口")
	flag.IntVar(&dstPort, "dstPort", 21, "目的端口")
}

func main() {
	var c api.IClient
	srcIP, dstIP, srcPort, dstPort := param()
	fmt.Println(srcIP, dstIP, srcPort, dstPort)
	c, err := client.NewGopacketFakeClient(srcIP, dstIP, srcPort, dstPort)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	tcp := layers.TCP{
		SrcPort: srcPort,
		DstPort: dstPort,
		Seq:     102,
		Ack:     103,
		SYN:     true,
		Window:  6666,
		Urgent: 1,
	}

	data := "hello is me"
	payload := gopacket.Payload([]byte(data))
	send, err := c.Send(&tcp, &payload)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("result len", send)
}

func param() (sIP, dIP net.IP, sPort, dPort layers.TCPPort) {
	sIP, _ = tools.String2IPV4(srcIP)
	dIP, _ = tools.String2IPV4(dstIP)
	sPort = tools.Int2TCPPort(srcPort)
	dPort = tools.Int2TCPPort(dstPort)
	return
}
