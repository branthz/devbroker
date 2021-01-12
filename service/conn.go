package service

import (
	"bufio"
	"net"
	"sync/atomic"
	"time"

	"github.com/branthz/devbroker/message"
	"github.com/branthz/devbroker/mqtt"
	"github.com/branthz/devbroker/topics"
	"github.com/branthz/utarrow/lib/log"
)

//Conn ...
type Conn struct {
	socket   net.Conn
	username string
	subs     *topics.Workq
	clientID string
	service  *Service
	route    *msgIndex
}

func (l *Listener) newConn(t net.Conn) *Conn {
	c := &Conn{
		socket:  t,
		route:   newMsgroute(),
		service: l.service,
		subs:    topics.New(),
	}
	atomic.AddInt64(&l.service.connections, 1)
	return c
}

//Close ...
func (c *Conn) Close() {}

//Process ...
func (c *Conn) Process() error {
	defer c.Close()
	reader := bufio.NewReaderSize(c.socket, 65535)
	maxSize := int64(20480)
	for {
		c.socket.SetReadDeadline(time.Now().Add(time.Second * 15))
		//TODO
		//限速
		msg, err := mqtt.DecodePacket(reader, maxSize)
		if err != nil {
			log.Errorln(err)
			return err
		}
		if err = c.onReceive(msg); err != nil {
			log.Errorln(err)
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
		for i, sub := range packet.Subscriptions {
			if err := c.bindSubscribe(sub.Topic); err != nil {
				ack.Qos[i] = 0x80
				continue
			}
			ack.Qos[i] = sub.Qos
		}
		log.Debug("publish-ack qos:%v---%d", ack.Qos, len(packet.Subscriptions))
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
		if packet.Header.QOS == 1 {
			ack := mqtt.Puback{MessageID: packet.MessageID}
			if _, err := ack.EncodeTo(c.socket); err != nil {
				return err
			}
		} else if packet.QOS == 2 {
			ack := mqtt.Pubrec{MessageID: packet.MessageID}
			if _, err := ack.EncodeTo(c.socket); err != nil {
				return err
			}
		}
	case mqtt.TypeOfPuback:
		//消费者返回的ack
		packet := msg.(*mqtt.Puback)
		if err := c.onPuback(packet.MessageID); err != nil {
			return err
		}
	case mqtt.TypeOfPubrel:
		packet := msg.(*mqtt.Pubrel)
		if err := c.onPubrelease(packet); err != nil {
			log.Errorln(err)
		}
		ack := mqtt.Pubcomp{MessageID: packet.MessageID}
		if _, err := ack.EncodeTo(c.socket); err != nil {
			return err
		}
		//
	}
	return nil
}

func (c *Conn) ID() string {
	return c.clientID
}

// Send forwards the message to the underlying client.
func (c *Conn) Send(dt []byte) (err error) {
	m, err := message.DecodeMessage(dt)
	if err != nil {
		return err
	}
	//defer c.MeasureElapsed("send.pub", time.Now())
	packet := mqtt.Publish{
		Header:  mqtt.Header{QOS: 0},
		Topic:   m.Topic,   // The channel for this message.
		Payload: m.Payload, // The payload for this message.
	}

	_, err = packet.EncodeTo(c.socket)
	return
}
