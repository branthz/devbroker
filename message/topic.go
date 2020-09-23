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
