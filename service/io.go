package service

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"time"
)

//Socket ...
type Socket struct {
	sync.RWMutex
	socket net.Conn
	writer bytes.Buffer
	reader sniffer
	cancel context.CancelFunc
}

func newSocket(c net.Conn) *Socket {
	s := &Socket{
		socket: c,
		reader: sniffer{source: c},
	}
	s.cancel = Repeat(context.Background(), time.Second, func() { s.Flush() })
	return s
}

//the pending buffer size.
func (s *Socket) Len() int {
	s.RLock()
	n := s.writer.Len()
	s.RUnlock()
	return n
}

func (s *Socket) Close() error {
	s.cancel()
	return s.socket.Close()
}

// Write writes the block of data into the underlying buffer.
func (s *Socket) Write(p []byte) (int, error) {

	// If we have reached the limit we can possibly write, queue up the packet.
	//if m.limit.Limit() {
	//	return m.enqueue(p)
	//}

	// If we have something in the buffer, flush everything.
	if s.Len() > 0 {
		s.enqueue(p)
		return s.Flush()
	}

	// Nothing in the buffer and we're not rate-limited, just write to the socket.
	return s.socket.Write(p)
}

func (s *Socket) Flush() (n int, err error) {
	if s.Len() == 0 {
		return 0, nil
	}

	// Flush everything and reset the buffer
	s.Lock()
	n, err = s.socket.Write(s.writer.Bytes())
	s.writer.Reset()
	s.Unlock()
	return
}

// LocalAddr returns the local network address.
func (s *Socket) LocalAddr() net.Addr {
	return s.socket.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (s *Socket) RemoteAddr() net.Addr {
	return s.socket.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
func (s *Socket) SetDeadline(t time.Time) error {
	return s.socket.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
func (s *Socket) SetReadDeadline(t time.Time) error {
	return s.socket.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
func (s *Socket) SetWriteDeadline(t time.Time) error {
	return s.socket.SetWriteDeadline(t)
}
func (s *Socket) enqueue(p []byte) (n int, err error) {
	s.Lock()
	n, err = s.writer.Write(p)
	s.Unlock()
	return
}

func (s *Socket) startSniffing() io.Reader {
	s.reader.reset(true)
	return &s.reader
}

func (s *Socket) doneSniffing() {
	s.reader.reset(false)
}

// Sniffer represents a io.Reader which can peek incoming bytes and reset back to normal.
type sniffer struct {
	source     io.Reader
	buffer     bytes.Buffer
	bufferRead int
	bufferSize int
	sniffing   bool
	lastErr    error
}

// Read reads data from the buffer.
func (s *sniffer) Read(p []byte) (int, error) {
	if s.bufferSize > s.bufferRead {
		bn := copy(p, s.buffer.Bytes()[s.bufferRead:s.bufferSize])
		s.bufferRead += bn
		return bn, s.lastErr
	} else if !s.sniffing && s.buffer.Cap() != 0 {
		s.buffer = bytes.Buffer{}
	}

	sn, sErr := s.source.Read(p)
	if sn > 0 && s.sniffing {
		s.lastErr = sErr
		if wn, wErr := s.buffer.Write(p[:sn]); wErr != nil {
			return wn, wErr
		}
	}
	return sn, sErr
}

// Reset resets the buffer.
func (s *sniffer) reset(snif bool) {
	s.sniffing = snif
	s.bufferRead = 0
	s.bufferSize = s.buffer.Len()
}
