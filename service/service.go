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

func (s *Service) Listen() error {
	//defer s.Close()
	ln, err := NewListen(config.GetConfig().Listen)
	if err != nil {
		return err
	}
	ln.Serve()
	return nil
}

//启动服务
func (s *Service) Run() error {
	err := s.Listen()
	if err != nil {
		return err
	}
	return nil
}
