package simpc

import (
	"net"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer(func(conn net.Conn, cominterface interface{}) {}, DefaultPath)
	if s == nil {
		t.Error("Faile to create server")
	}
}

func TestListen(t *testing.T) {
	s := NewServer(func(conn net.Conn, cominterface interface{}) {}, DefaultPath)
	err := s.Listen(0)
	if err != nil {
		t.Error("Failed to listen: " + err.Error())
	}
	defer s.Close()
}

func TestHandler(t *testing.T) {

}
