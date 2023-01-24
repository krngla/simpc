// simpc is a package that provides a simple IPC mechanism for Go programs, its usecase is quite specific, as such its capabilities are limited
package simpc

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/krngla/tmpfiler"
)

type server struct {
	ln       net.Listener
	done     chan struct{}
	handler  func(net.Conn, interface{})
	tempfile string
}

const (
	DefaultPath = "simpc_port.txt"
)

func NewServer(handler func(net.Conn, interface{}), tempfile string) *server {
	s := &server{}
	s.ln = nil
	s.done = make(chan struct{})
	s.handler = handler
	s.tempfile = tempfile
	return s
}

func (s server) PortStr() string {
	return strconv.Itoa(s.Port())
}

func (s server) Port() int {
	return s.ln.Addr().(*net.TCPAddr).Port
}

func (s *server) Listen(port int) error {
	var err error
	s.ln, err = net.Listen("tcp", ":"+s.PortStr())
	if err != nil {
		return errors.New("Failed to open port " + s.PortStr() + ": " + err.Error())
	}
	if port == 0 {
		_, err = tmpfiler.OpenWrite(s.tempfile, s.PortStr()+"\n")
	}
	if err != nil {
		return errors.New("Failed to write port to file: " + err.Error())
	}
	return nil
}

func (s *server) accept() (net.Conn, error) {
	return s.ln.Accept()
}

func (s *server) Dispatch(cominterface interface{}) error {
	if s.ln == nil {
		return errors.New("server not started")
	}
	for {
		select {
		case <-s.done:
			return nil
		default:
			conn, err := s.accept()
			if err != nil {
				return errors.New("failed to accept channel:" + err.Error())
			}
			go s.handler(conn, cominterface)
		}
	}
}

func (s *server) Close() {
	s.done <- struct{}{}
	s.ln.Close()
	tmpfiler.DeleteFile(s.tempfile)
}

type client struct {
	conn    net.Conn
	handler func(net.Conn, interface{})
}

func NewClient(port string, handler func(net.Conn, interface{})) *client {
	c := &client{}
	c.conn = nil
	c.handler = handler
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
		port, err = tmpfiler.OpenRead("MYIPC_port.txt")
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

func (c *client) Link(conIF interface{}) {
	go c.handler(c.conn, conIF)
}

func (c *client) Close() {
	c.conn.Close()
}
