package service

import (
	"github.com/branthz/devbroker/message"
)

func (c *Conn) onSubscribe(url []byte) error {
	ch := message.ParseTopic(url)

	return nil
}

func (c *Conn) onUnsubscribe(url []byte) error {
	return nil
}
