package service

import (
	"bufio"
	"net"
	"sync/atomic"
	"time"

	"github.com/branthz/devbroker/mqtt"
	"github.com/branthz/utarrow/lib/log"
)

//Conn ...
type Conn struct {
	socket   net.Conn
	username string
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
	return nil
}
