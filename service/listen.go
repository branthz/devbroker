package service

import (
	"net"

	"github.com/branthz/utarrow/lib/log"
)

type Listener struct {
	root    net.Listener
	service *Service
}

// Accept waits for and returns the next connection to the listener.
func (m *Listener) Accept() (net.Conn, error) {
	return m.root.Accept()
}

// Close closes the listener
func (m *Listener) Close() error {
	return m.root.Close()
}

// Addr returns the listener's network address.
func (m *Listener) Addr() net.Addr {
	return m.root.Addr()
}

func NewListen(addr string, ser *Service) (*Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Listener{
		root:    ln,
		service: ser,
	}, nil
}

func (m *Listener) Serve() error {
	for {
		c, err := m.root.Accept()
		if err != nil {
			log.Errorln(err)
			continue
		}
		go m.serve(c)
	}
}

func (m *Listener) serve(c net.Conn) {
	mc := m.newConn(c)
	mc.Process()
}
