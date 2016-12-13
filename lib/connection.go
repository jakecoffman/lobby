package lib

import (
	"golang.org/x/net/websocket"
)

// Connector wraps connections so tests are easy
type Connector interface {
	Send(v interface{}) error
	Recv(v interface{}) error
	Close() error
}

type WsConn struct {
	Conn *websocket.Conn
}

func (c *WsConn) Send(v interface{}) error {
	return websocket.JSON.Send(c.Conn, v)
}

func (c *WsConn) Recv(v interface{}) error {
	return websocket.JSON.Receive(c.Conn, v)
}

func (c *WsConn) Close() error {
	return c.Conn.Close()
}
