//wrapper of network handle
package service

import (
	"context"

	"github.com/branthz/devbroker/config"
	"github.com/branthz/devbroker/storage"
	"github.com/branthz/devbroker/topics"
)

type Service struct {
	context     context.Context
	tcp         *Server
	connections int64
	storage     storage.Storage
	tps         *topics.TopicPool
	//subscriptions *message.Trie
}

/*
func (s *Service) onAccept(t net.Conn) {
	conn := s.newConn(t)
	go conn.Process()
}*/

//构建实例
func NewService() (s *Service, err error) {
	s = &Service{
		tcp: new(Server),
	}
	//s.tcp.OnAccept = s.onAccept
	s.storage = storage.NewBolt(config.GetConfig().DataPath)
	s.tps = topics.New(s.storage)
	return
}

func (s *Service) Close() {

}

func (s *Service) Listen() {
	//defer s.Close()
	ln, err := NewListen(config.GetConfig().Listen, s)
	if err != nil {
		panic(err)
	}
	ln.Serve()
}

//启动服务
func (s *Service) Run() {
	s.Listen()
}

//-----------------------------------
