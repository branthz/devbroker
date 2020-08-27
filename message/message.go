package message

import (
	"sync"
)

type Counter struct {
	Channel []byte
	Ssid    Ssid
	Counter int
}

type Counters struct {
	sync.Mutex
	m map[uint32]*Counter
}

type Ssid []byte

func NewSsid(from []byte) Ssid {
	return from
}

type Message struct {
	ID      []byte
	topic   []byte
	Payload []byte
	TTL     uint32
}
