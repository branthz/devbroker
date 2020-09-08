package service

import (
	"github.com/branthz/devbroker/message"
	"github.com/branthz/devbroker/mqtt"
)

//url: clientid/topic
func (c *Conn) bindSubscribe(url []byte) error {
	ch := message.ParseTopic(string(url))
	c.subs.AddSub(string(ch.Topic), c)
	return nil
}

func (c *Conn) onUnsubscribe(url []byte) error {
	ch := message.ParseTopic(string(url))
	c.subs.UnSub(string(ch.Topic), c)
	return nil
}

func (c *Conn) onPublish(packet *mqtt.Publish) error {
	url := packet.Topic
	ch := message.ParseTopic(string(url))
	msg := message.NewMsg([]byte(ch.Id), []byte(ch.Topic), packet.Payload)
	c.service.storage.Store(msg)
	return nil
}
