package spyfall

import (
	"github.com/jakecoffman/lobby/lib"
	"gopkg.in/mgo.v2/bson"
)

type Spyfall struct {
	Id         string
	Code       string
	Players    []*lib.Player
	InProgress bool

	cmds chan *lib.PlayerCmd
}

const (
	START int = iota + 100
	STOP
)

func (s *Spyfall) Init() {
	s.Id = bson.NewObjectId().Hex()
	s.Code = "1234567"
	s.cmds = make(chan *lib.PlayerCmd)
	s.Players = []*lib.Player{}
}

func (s *Spyfall) Run(registry *lib.InMemoryRegistry) {
	for {
		cmd := <-s.cmds

		switch cmd.Type {
		case lib.NEW:
			fallthrough
		case lib.JOIN:
			s.Players = append(s.Players, cmd.Player)
		case lib.DISCONNECT:

		case lib.LEAVE:
			for i, p := range s.Players {
				if p.ID == cmd.Player.ID {
					s.Players = append(s.Players[:i], s.Players[i+1:]...)
					break
				}
			}
		case START:

		case STOP:

		default:
			continue
		}
		s.update()
	}
}

func (s *Spyfall) Send(cmd *lib.PlayerCmd) {
	s.cmds <- cmd
}

type state struct {
	Type    string
	Spyfall *Spyfall
}

func (s *Spyfall) update() {
	for _, p := range s.Players {
		_ = p.Send(state{Type: "spyfall", Spyfall: s})
		// ignoring errors because Player handles connection status
	}
}

func (s *Spyfall) String() string {
	return "Spyfall: " + s.Id
}

func (s *Spyfall) Location() string {
	return "spyfall/" + s.Id
}
