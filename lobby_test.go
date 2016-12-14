package lobby

import (
	"encoding/json"
	"fmt"
	"github.com/jakecoffman/lobby/games/spyfall"
	"github.com/jakecoffman/lobby/lib"
	"io"
	"log"
	"runtime"
	"testing"
	"time"
)

type MockConnection struct {
	id   string
	send chan []byte
	recv chan []byte
}

func NewMockConnection(id string) *MockConnection {
	return &MockConnection{
		id:   id,
		send: make(chan []byte, 10),
		recv: make(chan []byte, 10),
	}
}

// Send is called from Lobby only
func (c *MockConnection) Send(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.send <- data
	return nil
}

// Recv is called from Play only
func (c *MockConnection) Recv(v interface{}) error {
	data, ok := <-c.recv
	if !ok {
		return io.EOF
	}
	return json.Unmarshal(data, &v)
}

// used in test to receive something the server sent
func (c *MockConnection) clientSend(t *testing.T, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	timeout := time.After(1 * time.Second)
	select {
	case c.recv <- data:
		return nil
	case <-timeout:
		t.Fatal("Timed out writing")
	}
	return nil
}

// used in test to put something the server will handle
func (c *MockConnection) clientRecv(t *testing.T, v interface{}) error {
	timeout := time.After(1 * time.Second)

	select {
	case data := <-c.send:
		if err := json.Unmarshal(data, &v); err != nil {
			t.Fatal(err)
		}
	case <-timeout:
		_, file, line, _ := runtime.Caller(1)
		t.Fatal("Timedout reading at", fmt.Sprintf("%s:%d", file, line))
	}
	return nil
}

// Close is called on the server
func (c *MockConnection) Close() error {
	close(c.send)
	return nil
}

func (c *MockConnection) clientClose() {
	close(c.recv)
	<-c.send // Wait until lobby closes this
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type m map[string]interface{}

func TestWebSocketHandler(t *testing.T) {
	p1 := lib.NewPlayer()
	p2 := lib.NewPlayer()

	conn1 := NewMockConnection("1")
	conn2 := NewMockConnection("2")

	p1.Connect(conn1)
	p2.Connect(conn2)

	go Lobby.Play(p1)

	say := &say{}

	// Join

	conn1.clientRecv(t, say)
	if say.Message != fmt.Sprintf("Player %s joined", p1.GetName()) {
		t.Fatal(say.Message, "!= Player", p1.GetName(), "joined")
	}

	go Lobby.Play(p2)

	conn1.clientRecv(t, say)
	if say.Message != fmt.Sprintf("Player %s joined", p2.GetName()) {
		t.Fatal(say.Message, "!= Player", p2.GetName(), "joined")
	}
	conn2.clientRecv(t, say)
	if say.Message != fmt.Sprintf("Player %s joined", p2.GetName()) {
		t.Fatal(say.Message, "!= Player", p2.GetName(), "joined")
	}

	// Say

	msg := "Hello world!"
	conn2.clientSend(t, m{"type": SAY, "cmd": m{"Message": "Hello world!"}})

	conn1.clientRecv(t, say)
	expected := fmt.Sprintf("%s: %s", p2.GetName(), msg)
	if say.Message != expected {
		t.Fatal(say.Message, "!=", expected)
	}
	conn2.clientRecv(t, say)
	if say.Message != expected {
		t.Fatal(say.Message, "!=", expected)
	}

	// Rename
	previousName := p1.GetName()
	conn1.clientSend(t, m{"type": NAME, "cmd": m{"Message": "Bob"}})

	conn1.clientRecv(t, say)
	expected = fmt.Sprintf("%s is now known as %s", previousName, p1.GetName())
	if say.Message != expected {
		t.Fatal(say.Message, "!=", expected)
	}
	conn2.clientRecv(t, say)
	if say.Message != expected {
		t.Fatal(say.Message, "!=", expected)
	}
	if p1.GetName() != "Bob" {
		t.Fatal(p1.GetName(), "!=", "Bob")
	}

	// Leave
	conn1.clientClose()

	conn2.clientRecv(t, say)
	expected = fmt.Sprintf("Player %s left", p1.GetName())
	if say.Message != expected {
		t.Fatal(say.Message, "!=", expected)
	}

	conn2.clientClose()
}
