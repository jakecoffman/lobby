package lobby

import (
	"encoding/json"
	"fmt"
	"github.com/jakecoffman/lobby/lib"
	"io"
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
		simple := &lib.SimpleCmd{}
		if err := json.Unmarshal(cmd.Command, simple); err != nil {
			log.Println(err)
			continue
		}

		switch cmd.Type {
		case lib.JOIN:
			l.players[simple.Player.ID] = simple.Player
			l.broadcast(fmt.Sprintf("%s: joined", simple.Player.GetName()))
		case lib.DISCONNECT:
			l.broadcast(fmt.Sprintf("%s: disconnected", simple.Player.GetName()))
		case lib.LEAVE:
			delete(l.players, simple.Player.ID)
			l.broadcast(fmt.Sprintf("%s: left", simple.Player.GetName()))
		case lib.SAY: // TODO remove from lib
			l.broadcast(fmt.Sprintf("%s: %s", simple.Player.GetName(), simple.Message))
		}
	}
}

func (l *Lobby) Send(cmd *lib.PlayerCmd) {
	l.cmds <- cmd
}

func (l *Lobby) broadcast(msg string) {
	for _, p := range l.players {
		if err := p.Send(say{Type: "SAY", Msg: msg}); err != nil {
			if err != io.EOF {
				log.Println(err)
			}
		}
	}
}
