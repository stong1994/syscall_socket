package tools

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/google/gopacket/layers"
	"net"
)

var (
	ErrIPNotValid = errors.New("ip is not valid")
)

func String2IPV4(ip string) (net.IP, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return nil, ErrIPNotValid
	}
	return ipv4, nil
}

func Int2TCPPort(port int) layers.TCPPort {
	return layers.TCPPort(uint16(port))
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
