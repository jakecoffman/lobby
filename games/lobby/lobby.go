package lobby

import (
	"fmt"
	"github.com/jakecoffman/lobby/lib"
	"log"
)

// Lobby is a chat room game, where you chat!
type Lobby struct {
	players map[string]*lib.Player

	cmds chan *lib.PlayerCmd
}

type say struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

func (l *Lobby) Init() {
	l.players = map[string]*lib.Player{}
	l.cmds = make(chan *lib.PlayerCmd)
}

func (l *Lobby) Run(registry *lib.InMemoryRegistry) {
	for {
		cmd := <-l.cmds

		switch cmd.Type {
		case lib.JOIN:
			l.players[cmd.Player.ID] = cmd.Player
			l.broadcast(fmt.Sprintf("%s: joined", cmd.Player.GetName()))
		case lib.DISCONNECT:
			l.broadcast(fmt.Sprintf("%s: disconnected", cmd.Player.GetName()))
		case lib.LEAVE:
			delete(l.players, cmd.Player.ID)
			l.broadcast(fmt.Sprintf("%s: left", cmd.Player.GetName()))
		case lib.SAY: // TODO remove from lib
			if simple, err := cmd.SimpleCmd(); err != nil {
				log.Println(err)
				continue
			} else {
				l.broadcast(fmt.Sprintf("%s: %s", cmd.Player.GetName(), simple))
			}
		}
	}
}

func (l *Lobby) Send(cmd *lib.PlayerCmd) {
	l.cmds <- cmd
}

func (l *Lobby) broadcast(msg string) {
	for _, p := range l.players {
		_ = p.Send(&say{Type: "say", Msg: msg})
		// ignoring errors because Player handles connection status
	}
}
