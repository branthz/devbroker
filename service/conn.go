package service

import (
	"bufio"
	"net"
	"sync/atomic"
	"time"

	"github.com/branthz/devbroker/message"
	"github.com/branthz/devbroker/mqtt"
	"github.com/branthz/utarrow/lib/log"
)

//Conn ...
type Conn struct {
	socket   net.Conn
	username string
	subs     *message.SubContainer
	clientID string
	service  *Service
}

func newConn(c net.Conn) *Conn {
	return &Conn{
		socket: c,
	}
}

func (s *Service) newConn(t net.Conn) *Conn {
	c := &Conn{
		socket: t,
	}
	atomic.AddInt64(&s.connections, 1)
	return c
}

//Close ...
func (c *Conn) Close() {}

//Process ...
func (c *Conn) Process() error {
	defer c.Close()
	reader := bufio.NewReaderSize(c.socket, 65535)
	maxSize := int64(2048)
	for {
		c.socket.SetDeadline(time.Now().Add(time.Second * 10))
		//TODO
		//限速
		msg, err := mqtt.DecodePacket(reader, maxSize)
		if err != nil {
			return err
		}
		if err = c.onReceive(msg); err != nil {
			return err
		}
	}
	return nil
}

func (c *Conn) onReceive(msg mqtt.Message) error {
	log.Infoln("receive,type:", msg.Type(), "data:", msg.String())
	switch msg.Type() {
	//客户端链接
	//响应
	case mqtt.TypeOfConnect:
		var result uint8
		ack := mqtt.Connack{ReturnCode: result}
		if _, err := ack.EncodeTo(c.socket); err != nil {
			return err
		}
	//订阅
	//可以一次多个频道
	//
	case mqtt.TypeOfSubscribe:
		packet := msg.(*mqtt.Subscribe)
		ack := mqtt.Suback{
			MessageID: packet.MessageID,
			Qos:       make([]uint8, len(packet.Subscriptions)),
		}
		for _, sub := range packet.Subscriptions {
			if err := c.bindSubscribe(sub.Topic); err != nil {
				ack.Qos = append(ack.Qos, 0x80)
				continue
			}
			ack.Qos = append(ack.Qos, sub.Qos)
		}
		if _, err := ack.EncodeTo(c.socket); err != nil {
			return err
		}

	case mqtt.TypeOfUnsubscribe:
		packet := msg.(*mqtt.Unsubscribe)
		ack := mqtt.Suback{
			MessageID: packet.MessageID,
		}
		for _, sub := range packet.Topics {
			if err := c.onUnsubscribe(sub.Topic); err != nil {
				//notify err
			}
		}
		if _, err := ack.EncodeTo(c.socket); err != nil {
			return err
		}
	case mqtt.TypeOfPingreq:
		ack := mqtt.Pingresp{}
		if _, err := ack.EncodeTo(c.socket); err != nil {
			return err
		}
	case mqtt.TypeOfDisconnect:
		return nil
	case mqtt.TypeOfPublish:
		packet := msg.(*mqtt.Publish)
		if err := c.onPublish(packet); err != nil {
			log.Errorln(err)
			// c.notifyError(err, packet.MessageID)
		}

		// Acknowledge the publication
		if packet.Header.QOS > 0 {
			ack := mqtt.Puback{MessageID: packet.MessageID}
			if _, err := ack.EncodeTo(c.socket); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Conn) ID() string {
	return c.clientID
}

func (c *Conn) Send(m *message.Message) error {
	return nil
}
