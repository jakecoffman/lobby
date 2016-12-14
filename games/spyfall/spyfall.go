package spyfall

import (
	"encoding/json"
	"github.com/jakecoffman/lobby/lib"
	"gopkg.in/mgo.v2/bson"
	"log"
)

// TODO: in main call lobby Register with Spyfall instance

type Spyfall struct {
	id      string
	players map[string]*lib.Player

	join       chan *lib.Player
	rejoin     chan *lib.Player
	leave      chan *lib.Player
	disconnect chan *lib.Player
	msgs       chan *json.RawMessage
}

// this is the server
func (s *Spyfall) Run(l lib.Lobbyer) {
	l.RunCallback(s.id, s)
	s.players = map[string]*lib.Player{}
	s.join = make(chan *lib.Player)
	s.leave = make(chan *lib.Player)
	s.disconnect = make(chan *lib.Player)
	s.msgs = make(chan *json.RawMessage)
	s.id = bson.NewObjectId().Hex()
	for {
		select {
		case player := <-s.join:
			s.players[player.ID] = player
			player.GameId = s.id
		case player := <-s.rejoin:
			s.players[player.ID] = player
			player.GameId = s.id
		case player := <-s.disconnect:
			log.Println(player, "disconnected")
		case player := <-s.leave:
			delete(s.players, player.ID)
			player.GameId = ""
		case _ = <-s.msgs:
			//log.Println()
		}
	}
}

// Everything below is the client's job to send
func (s *Spyfall) Join(player *lib.Player) {
	s.join <- player
}

func (s *Spyfall) Rejoin(player *lib.Player) {
	s.rejoin <- player
}

func (s *Spyfall) Disconnect(player *lib.Player) {
	s.disconnect <- player
}

func (s *Spyfall) Leave(player *lib.Player) {
	s.leave <- player
}

func (s *Spyfall) Send(data *json.RawMessage) {
	s.msgs <- data
}
