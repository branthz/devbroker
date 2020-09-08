package message

import "strings"

type channel struct {
	Id    string
	Topic string
}

func ParseTopic(src string) *channel {
	dd := strings.Split(src, "/")
	c := new(channel)
	c.Id = dd[0]
	c.Topic = dd[1]
	return c
}

type Subscriber interface {
	ID() string
	Send(*Message) error
}

type SubContainer struct {
	cn map[string][]Subscriber
}

func (s *SubContainer) AddSub(topic string, con Subscriber) {
	if v, ok := s.cn[topic]; ok {
		v = append(v, con)
		s.cn[topic] = v
	} else {
		s.cn[topic] = []Subscriber{con}
	}
}

func (s *SubContainer) UnSub(topic string, con Subscriber) {
	if v, ok := s.cn[topic]; ok {
		if len(v) > 1 {
			re := make([]Subscriber, 0)
			for _, vv := range v {
				if vv.ID() == con.ID() {
					continue
				}
				re = append(re, vv)
			}
			s.cn[topic] = re
		} else {
			delete(s.cn, topic)
		}
	}
}
