package api

import "github.com/google/gopacket"

type IClient interface {
	Send(transferLayer, payload gopacket.SerializableLayer) (int, error)
	Close()
}
