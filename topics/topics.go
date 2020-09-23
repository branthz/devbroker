package topics

import (
	"github.com/branthz/devbroker/message"
)

type Subscriber interface {
	ID() string
	Send(*message.Message) error
}

//这种定义只支持单一消费者模式，key为主题,想要广播效果就是多个消费者，可将key扩展成主题+频道
type Workq struct {
	cn map[string]Subscriber
}

func (s *Workq) AddSub(topic string, con Subscriber) {
	s.cn[topic] = con
}

func (s *Workq) UnSub(topic string, con Subscriber) {
	delete(s.cn, topic)
}
