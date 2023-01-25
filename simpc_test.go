package simpc

import (
	"net"
	"os/exec"
	"strings"
	"testing"
)

type mockServerHandler struct {
	ran  bool
	nran int
}

func (m *mockServerHandler) Handle(conn net.Conn) {
	m.ran = true
	m.nran++
}

func TestNewServer(t *testing.T) {
	s := NewServer(nil, DefaultPath, nil)
	if s == nil {
		t.Error("Failed to create server")
	}
}

func TestListen(t *testing.T) {
	s := NewServer(nil, DefaultPath, nil)
	err := s.Listen(0)
	if err != nil {
		t.Error("Failed to listen: " + err.Error())
	}
	defer s.Close()

	cmd := exec.Command("cmd", "/C", "netstat", "-an", "|", "findstr", "0.0.0.0:"+s.PortStr())

	ret, err := cmd.Output()
	if err != nil {
		t.Errorf("failed to execute command %q: %v", cmd.String(), err.Error())
	}
	cmd.Wait()
	if !strings.Contains(string(ret), s.PortStr()) {
		t.Error("Port not found")
	}
	if !strings.Contains(string(ret), "LISTENING") {
		t.Error("Port not listening")
	}

}

func TestNewClient(t *testing.T) {
	c := NewClient(nil, DefaultPath)

	if c == nil {
		t.Error("Failed to create client")
	}

}
func TestConnect(t *testing.T) {
	s := NewServer(nil, DefaultPath, nil)
	_ = s.Listen(0)
	defer s.Close()

	c := NewClient(nil, DefaultPath)
	err := c.Connect(s.PortStr())
	if err != nil {
		t.Error("Failed to dial: " + err.Error())
	}
	defer c.Close()

}

type mockIO struct {
	ran bool
}

func (m *mockIO) Handle(conn net.Conn) {
	m.ran = true
	conn.Close()
}

func TestServerHandler(t *testing.T) {

	m := &mockIO{false}
	s := NewServer(m, DefaultPath, nil)
	s.handler.Handle(nil)
	if !m.ran {
		t.Error("Failed to set handler")
	}
}

func TestClientHandler(t *testing.T) {

	m := &mockIO{false}
	c := NewClient(m, DefaultPath)
	c.handler.Handle(nil)
	if !m.ran {
		t.Error("Failed to set handler")
	}
}

func launchdispatch(s *server, errs chan error) {
	err := s.Dispatch()
	if err != nil {
		errs <- err
	}
	errs <- nil
}

func TestServerDispatch(t *testing.T) {
	done := make(chan struct{})
	ms := &mockServerHandler{false, 0}
	s := NewServer(ms, DefaultPath, done)
	err := s.Listen(0)
	if err != nil {
		t.Error("Failed to listen: " + err.Error())
	}
	errs := make(chan error)
	go launchdispatch(s, errs)

	mc := &mockIO{false}
	c := NewClient(mc, DefaultPath)

	_ = c.Connect(s.PortStr())
	c.Close()
	c.Link()
	go func() {
		done <- struct{}{}
	}()
	_ = c.Connect(s.PortStr())
	c.Close()

	err = <-errs
	if err != nil {
		t.Error("Dispatch failed: " + err.Error())
	}
	if !ms.ran {
		t.Error("Dispatch failed to run handler")
	}
	if ms.nran != 2 {
		t.Error("Dispatch failed to run handler twice")
	}
}

func TestServerMassDispatch(t *testing.T) {
	n := 10_000
	done := make(chan struct{})
	ms := &mockServerHandler{false, 0}
	s := NewServer(ms, DefaultPath, done)
	s.Listen(0)
	errs := make(chan error)
	go launchdispatch(s, errs)

	clientfactory := func() {
		c := NewClient(nil, DefaultPath)
		_ = c.Connect(s.PortStr())
		c.Close()
	}

	for i := 0; i < n-1; i++ {
		go clientfactory()
	}
	go func() {
		done <- struct{}{}
	}()
	clientfactory()

	err := <-errs
	if err != nil {
		t.Error("Dispatch failed: " + err.Error())
	}
	if !ms.ran {
		t.Error("Dispatch failed to run handler")
	}
	if ms.nran != n {
		t.Errorf("Dispatch failed to run handler %d times", n)
	}
}
