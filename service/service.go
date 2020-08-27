//wrapper of network handle
package service

import (
	"context"
	"net"

	"github.com/branthz/devbroker/config"
)

type Service struct {
	context     context.Context
	tcp         *Server
	connections int64
	//subscriptions *message.Trie
}

func (s *Service) onAccept(t net.Conn) {
	conn := s.newConn(t)
	go conn.Process()
}

//构建实例
func NewService() (s *Service, err error) {
	s = &Service{
		tcp: new(Server),
	}
	s.tcp.OnAccept = s.onAccept
	return
}

func (s *Service) Close() {

}

func (s *Service) Listen() {
	//defer s.Close()
	ln, err := NewListen(config.GetConfig().Listen)
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
