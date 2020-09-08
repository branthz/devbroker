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
	m map[string]*Counter
}

func (c *Counters) Bind(topic []byte, cid []byte) {
	c.Lock()
	defer c.Unlock()
	//TODO
	//if exist return value
	c.m[string(topic)] = &Counter{
		Channel: topic,
		Ssid:    cid,
		Counter: 1,
	}
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

func NewMsg(id []byte, topic []byte, dt []byte) *Message {
	m := new(Message)
	m.topic = topic
	m.Payload = dt
	m.ID = id
	return m
}
