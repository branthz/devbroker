package service

import (
	"context"
	"net"
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
	defer s.Close()
	//ln, err := net.ListenTCP("127.0.0.1:1234", nil)
	//if err != nil {
	//	return err
	//}

	return nil
}
