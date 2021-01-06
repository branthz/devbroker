package service

import (
	"github.com/branthz/devbroker/message"
	"github.com/branthz/devbroker/mqtt"
)

//url: clientid/topic
func (c *Conn) bindSubscribe(url []byte) error {
	ch := message.ParseTopic(string(url))
	c.subs.AddSub(string(ch.Topic), c, c.service.storage)
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
	c.service.storage.SaveMsg(ch.Topic, msg.Encode())
	return nil
}

func (c *Conn) onPuback(id uint16) error {
	//TODO
	//消费的消息确认，read commit
	//现在id怎么和存储里的fileid+offset关联
	//这种存储是易失性的，conn重建后丢失;有空替换掉，
	path := c.route.getroute(id)
	topic := "abc"
	c.service.storage.CommitRead(topic, path)
	return nil
}
