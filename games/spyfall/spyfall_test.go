package spyfall

import (
	"github.com/jakecoffman/lobby/lib"
	"github.com/jakecoffman/lobby/mocks"
	"testing"
)

func TestSpyfall(t *testing.T) {
	msg := &state{}

	p1 := &lib.User{ID: "1", Name: "P1"}
	p2 := &lib.User{ID: "2", Name: "P2"}

	conn1 := mocks.NewMockConn()
	conn2 := mocks.NewMockConn()

	p1.Connect(conn1)
	p2.Connect(conn2)

	s := Spyfall{}
	s.Init("1", "code")

	go s.Run()

	s.Send(lib.NewPlayerCmd("JOIN", p1))
	conn1.TestRecv(msg)

	s.Send(lib.NewPlayerCmd("JOIN", p2))
	conn1.TestRecv(msg)
	conn2.TestRecv(msg)

	s.Send(lib.NewPlayerCmd("READY", p1))
	conn1.TestRecv(msg)
	conn2.TestRecv(msg)

	players := msg.Spyfall.Players
	if players[0].Ready != true && players[1].Ready != true {
		t.Fatal("Neither player sent their correct readyness")
	}

	s.Send(lib.NewPlayerCmd("LEAVE", p1))
	conn2.TestRecv(msg)

	s.Send(lib.NewPlayerCmd("LEAVE", p2))

}
