package spyfall

import (
	"encoding/json"
	"github.com/jakecoffman/lobby/lib"
	"log"
	"testing"
)

type MockRegistry struct{}

func (m *MockRegistry) Register(lib.Game, string) {

}

func (m *MockRegistry) Start(string) (lib.Game, error) {
	return nil, nil
}

func (m *MockRegistry) Find(string) (lib.Game, error) {
	return nil, nil
}

type MockConn struct {
	send chan []byte
	recv chan []byte
}

func NewMockConn() *MockConn {
	return &MockConn{
		send: make(chan []byte, 3),
		recv: make(chan []byte, 3),
	}
}

func (c *MockConn) Send(v interface{}) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.send <- bytes
	return nil
}

func (c *MockConn) Recv(v interface{}) error {
	bytes := <-c.recv
	return json.Unmarshal(bytes, v)
}

func (c *MockConn) Close() error {
	log.Println("TODO: handle close")
	return nil
}

func (c *MockConn) TestSend(v interface{}) {
	if bytes, err := json.Marshal(v); err != nil {
		panic(err)
	} else {
		c.recv <- bytes
	}
}

func (c *MockConn) TestRecv(v interface{}) {
	bytes := <-c.send
	if err := json.Unmarshal(bytes, v); err != nil {
		panic(err)
	}
}

func TestSpyfall(t *testing.T) {
	msg := &state{}

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
	conn1.TestRecv(msg)

	s.Send(lib.NewSimpleCmd("JOIN", "", p2))
	conn1.TestRecv(msg)
	conn2.TestRecv(msg)

	s.Send(lib.NewSimpleCmd("READY", "", p1))
	conn1.TestRecv(msg)
	conn2.TestRecv(msg)

	players := msg.Spyfall.Players
	if players[0].Ready != true && players[1].Ready != true {
		t.Fatal("Neither player sent their correct readyness")
	}

	s.Send(lib.NewSimpleCmd("LEAVE", "", p1))
	conn2.TestRecv(msg)

	s.Send(lib.NewSimpleCmd("LEAVE", "", p2))

}
