package service

import (
	"net"

	"github.com/branthz/utarrow/lib/log"
)

type Listener struct {
	root net.Listener
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

func NewListen(addr string) (*Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Listener{
		root: ln,
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
	mc := newConn(c)
	mc.Process()
}
