package message

import (
	"bytes"
	"sync"

	"github.com/golang/snappy"
	"github.com/kelindar/binary"
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
	ID      uint64
	Topic   []byte
	Payload []byte
	TTL     uint32
}

func NewMsg(id uint64, topic []byte, dt []byte) *Message {
	m := new(Message)
	m.Topic = topic
	m.Payload = dt
	m.ID = id
	return m
}

func (m *Message) Encode() []byte {
	encoder := encoders.Get().(*binary.Encoder)
	defer encoders.Put(encoder)

	buffer := encoder.Buffer().(*bytes.Buffer)
	buffer.Reset()

	// Encode into a temporary buffer
	if err := encoder.Encode(m); err != nil {
		panic(err) // Should never panic
	}

	// Decode from snappy with an allocation done by providing 'nil' destination.
	return snappy.Encode(nil, buffer.Bytes())
}

func DecodeMessage(buf []byte) (out Message, err error) {
	// We need to allocate, given that the unmarshal is now no-copy. By using 'nil' as destination
	// we make sure that the underlying buffer is calculated based on the decoded length.
	if buf, err = snappy.Decode(nil, buf); err == nil {
		err = binary.Unmarshal(buf, &out)
	}
	return
}
