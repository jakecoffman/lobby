package mocks

import (
	"encoding/json"
	"log"
)

type MockConn struct {
	send chan []byte
	recv chan []byte
}

func NewMockConn() *MockConn {
	return &MockConn{
		send: make(chan []byte),
		recv: make(chan []byte),
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
