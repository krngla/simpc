// simpc is a package that provides a simple IPC mechanism for Go programs, its usecase is quite specific, as such its capabilities are limited
package simpc

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/krngla/tmpfiler"
)

type ServerHandler interface {
	Handle(net.Conn)
}

type server struct {
	ln        net.Listener
	Listening bool
	done      <-chan struct{}
	handler   ServerHandler
	tempfile  string
}

const (
	DefaultPath = "simpc_port.txt"
)

func NewServer(handler ServerHandler, tempfile string, done <-chan struct{}) *server {
	s := &server{}
	s.ln = nil
	s.Listening = false
	s.done = done
	s.handler = handler
	s.tempfile = tempfile
	return s
}

func (s server) PortStr() string {
	if s.ln == nil {
		return "-1"
	}
	return strconv.Itoa(s.Port())
}

func (s server) Port() int {
	if s.ln == nil {
		return -1
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

func (s *server) Listen(port int) error {
	var err error
	s.ln, err = net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return errors.New("Failed to open port " + s.PortStr() + ": " + err.Error())
	}
	if port != 0 {
		return nil
	}
	_, err = tmpfiler.OpenWrite(s.tempfile, s.PortStr()+"\n")
	if err != nil {
		return errors.New("Failed to write port to file: " + err.Error())
	}
	return nil
}

func (s *server) accept() (net.Conn, error) {
	if s.ln == nil {
		return nil, errors.New("server not started")
	}
	return s.ln.Accept()
}

func (s *server) Dispatch() error {
	if s.ln == nil {
		return errors.New("server not started")
	}
	s.Listening = true

	for {
		select {
		case <-s.done:
			s.Listening = false
			return nil
		default:
			conn, err := s.accept()
			if err != nil {
				s.Listening = false
				return errors.New("failed to accept channel:" + err.Error())
			}
			go s.handler.Handle(conn)
		}
	}
}

func (s *server) Close() {

	if s.ln != nil {
		s.ln.Close()
	}
	tmpfiler.DeleteFile(s.tempfile)
}

type clientHandler interface {
	Handle(conn net.Conn)
}

type client struct {
	conn     net.Conn
	handler  clientHandler
	tempfile string
}

func NewClient(handler clientHandler, tempfile string) *client {
	c := &client{}
	c.conn = nil
	c.handler = handler
	c.tempfile = tempfile
	return c
}

func (c *client) dial(port string) error {
	var err error
	c.conn, err = net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return errors.New("failed to connect:" + err.Error())
	}
	return nil
}

func (c *client) Connect(port string) error {
	if port == "0" {
		var err error
		port, err = tmpfiler.OpenRead(c.tempfile)
		if err != nil {
			return errors.New("failed to read port configuration file:" + err.Error())
		}
		port = strings.TrimSuffix(port, "\n")
	}

	err := c.dial(port)
	if err != nil {
		return errors.New("failed to connect:" + err.Error())
	}
	return nil
}

func (c *client) Link() {
	c.handler.Handle(c.conn)
}

func (c *client) Close() {
	if c.conn == nil {
		return
	}
	c.conn.Close()
}
