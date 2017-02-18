package spyfall

import (
	"github.com/jakecoffman/lobby/lib"
	"github.com/jakecoffman/lobby/mocks"
	"os"
	"runtime/pprof"
	"testing"
	"time"
)

func TestSpyfall(t *testing.T) {
	// make testing faster
	countdownDuration = 0

	go func() {
		<-time.After(8 * time.Second)
		pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
		panic("Took too long")
	}()

	stateMsg1 := &state{}
	stateMsg2 := &state{}
	simpleMsg := &lib.SimpleMsg{}

	p1 := &lib.User{ID: "1", Name: "P1"}
	p2 := &lib.User{ID: "2", Name: "P2"}

	conn1 := mocks.NewMockConn()
	conn2 := mocks.NewMockConn()

	p1.Connect(conn1)
	p2.Connect(conn2)

	s := &Spyfall{}
	s.Init("1", "code")

	c := make(chan struct{})

	go func() {
		s.Run()
		c <- struct{}{}
	}()

	s.Send(lib.NewPlayerCmd("JOIN", p1))
	conn1.TestRecv(stateMsg1)

	s.Send(lib.NewPlayerCmd("JOIN", p2))
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg1)

	s.Send(lib.NewPlayerCmd("READY", p1))
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg1)

	players := stateMsg1.Spyfall.Players
	if players[0].Ready == false && players[1].Ready == false {
		t.Fatal("One player should be ready at this point", stateMsg1)
	}

	s.Send(lib.NewPlayerCmd("READY", p2))
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg1)
	players = stateMsg1.Spyfall.Players
	if players[0].Ready == false || players[1].Ready == false {
		t.Fatal("Both players should be ready at this point")
	}

	// game starting
	conn1.TestRecv(simpleMsg)
	conn2.TestRecv(simpleMsg)
	if simpleMsg.Type != "starting" && simpleMsg.Msg != "Game starts in 3" {
		t.Fatal("Wrong starting message", simpleMsg.Type, simpleMsg.Msg)
	}
	// 2
	conn1.TestRecv(simpleMsg)
	conn2.TestRecv(simpleMsg)
	// 1
	conn1.TestRecv(simpleMsg)
	conn2.TestRecv(simpleMsg)

	// the actual game started
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg2)

	if !s.Started || !stateMsg1.Spyfall.Started || !stateMsg2.Spyfall.Started {
		t.Fatal("Game didn't start")
	}

	if stateMsg1.You.Spy == false && stateMsg2.You.Spy == false {
		t.Fatal("Spy was not assigned")
	}

	s.Send(lib.NewPlayerCmd("DISCONNECT", p2))
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg2)

	s.Send(lib.NewPlayerCmd("JOIN", p2))
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg1)

	s.Send(lib.NewPlayerCmd("STOP", p1))
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg1)

	s.Send(lib.NewPlayerCmd("STOP", p2))
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg1)

	if stateMsg1.Spyfall.Started != false || s.Started != false {
		t.Fatal("Game should be stopped")
	}

	// restart the game
	s.Send(lib.NewPlayerCmd("READY", p1))
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg1)

	players = stateMsg1.Spyfall.Players
	if players[0].Ready == false && players[1].Ready == false {
		t.Fatal("One player should be ready at this point", stateMsg1)
	}

	s.Send(lib.NewPlayerCmd("READY", p2))
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg1)
	players = stateMsg1.Spyfall.Players
	if players[0].Ready == false || players[1].Ready == false {
		t.Fatal("Both players should be ready at this point")
	}

	// game starting
	conn1.TestRecv(simpleMsg)
	conn2.TestRecv(simpleMsg)
	if simpleMsg.Type != "starting" && simpleMsg.Msg != "Game starts in 3" {
		t.Fatal("Wrong starting message", simpleMsg.Type, simpleMsg.Msg)
	}
	// 2
	conn1.TestRecv(simpleMsg)
	conn2.TestRecv(simpleMsg)
	// 1
	conn1.TestRecv(simpleMsg)
	conn2.TestRecv(simpleMsg)

	// the actual game started
	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg2)

	if !s.Started || !stateMsg1.Spyfall.Started || !stateMsg2.Spyfall.Started {
		t.Fatal("Game didn't start")
	}

	// force the game to end
	s.timer.Reset(1 * time.Microsecond)

	conn1.TestRecv(stateMsg1)
	conn2.TestRecv(stateMsg2)
	if stateMsg1.Spyfall.Started || stateMsg2.Spyfall.Started {
		t.Fatal("Game should have ended")
	}

	s.Send(lib.NewPlayerCmd("LEAVE", p1))
	conn2.TestRecv(stateMsg1)
	if len(stateMsg1.Spyfall.Players) != 1 {
		t.Fatal("Players should but 1 but is", len(stateMsg1.Spyfall.Players))
	}

	s.Send(lib.NewPlayerCmd("LEAVE", p2))
	<-c
	if len(s.Players) != 0 {
		t.Fatal("Players should be 0 but is", len(s.Players))
	}

	// TODO: test to make sure the game removes it's "code" when all players leave

}
