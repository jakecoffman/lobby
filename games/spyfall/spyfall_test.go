package spyfall

import (
	"testing"
	"github.com/jakecoffman/lobby/lib"
	"log"
)

type MockRegistry struct {}

func (m *MockRegistry) Register(lib.Game, string) {

}

func (m *MockRegistry) Start(string) (lib.Game, error) {
	return nil, nil
}

func (m *MockRegistry) Find(string) (lib.Game, error) {
	return nil, nil
}

type MockConn struct {
	send chan interface{}
	recv chan interface{}
}

func NewMockConn() *MockConn {
	return &MockConn{
		send: make(chan interface{}, 3),
		recv: make(chan interface{}, 3),
	}
}

func (c *MockConn) Send(v interface{}) error {
	c.send <- v
	return nil
}

func (c *MockConn) Recv(v interface{}) error {
	v = <- c.recv
	return nil
}

func (c *MockConn) Close() error {
	log.Println("TODO: handle close")
	return nil
}

func (c *MockConn) TestSend(v interface{}) {
	c.recv <- v
}

func (c *MockConn) TestRecv() (interface{}, bool) {
	v, ok := <- c.send
	return v, ok
}

func TestSpyfall(t *testing.T) {
	var msg interface{}

	p1 := &lib.User{ID: "1", Name: "P1"}
	p2 := &lib.User{ID: "2", Name: "P2"}

	conn1 := NewMockConn()
	conn2 := NewMockConn()

	p1.Connect(conn1, &MockRegistry{})
	p2.Connect(conn2, &MockRegistry{})

	s := Spyfall{}
	s.Init("1", "code")

	go s.Run()

	s.Send(lib.NewSimpleCmd("JOIN", "", p1))
	msg, _ = conn1.TestRecv()
	log.Printf("%#v\n", msg.(state).Spyfall)

	s.Send(lib.NewSimpleCmd("JOIN", "", p2))
	msg, _ = conn1.TestRecv()
	log.Println(msg)
	msg, _ = conn2.TestRecv()
	log.Println(msg)

	s.Send(lib.NewSimpleCmd("LEAVE", "", p1))
	msg, _ = conn2.TestRecv()
	log.Println(msg)

	s.Send(lib.NewSimpleCmd("LEAVE", "", p2))

}
