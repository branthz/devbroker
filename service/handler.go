package service

import (
	"github.com/branthz/devbroker/message"
	"github.com/branthz/devbroker/mqtt"
)

//url: clientid/topic
func (c *Conn) bindSubscribe(url []byte) error {
	//ch := message.ParseTopic(string(url))
	c.subs.AddSub(string(url), c, c.service.storage)
	return nil
}

func (c *Conn) onUnsubscribe(url []byte) error {
	ch := message.ParseTopic(string(url))
	c.subs.UnSub(string(ch.Topic), c)
	return nil
}

func (c *Conn) onPublish(packet *mqtt.Publish) error {
	//url := packet.Topic
	//ch := message.ParseTopic(string(url))
	msg := message.NewMsg(uint64(packet.MessageID), []byte(packet.Topic), packet.Payload)
	if packet.QOS == 2 {
		pid := string(packet.MessageID) + "/" + "hello"
		c.service.storage.PreSaveMsg([]byte(pid), msg.Encode())
	}
	c.service.storage.SaveMsg(string(packet.Topic), msg.Encode())
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

func (c *Conn) onPubrelease(packet *mqtt.Pubrel) error {
	pid := string(packet.MessageID) + "/" + "hello"
	err := c.service.storage.CommitMsg([]byte(pid))
	return err
}
