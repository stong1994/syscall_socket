package api

import "github.com/google/gopacket"

type IClient interface {
	Send(layer gopacket.SerializableLayer) (int, error)
	Close()
}
