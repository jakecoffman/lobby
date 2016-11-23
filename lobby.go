package lobby

import (
	"fmt"
	"io"
	"log"

	"github.com/jakecoffman/lobby/lib"
)

var Lobby *lobby

func init() {
	// lobby is a singleton
	Lobby = NewLobby()
	go Lobby.run()
}

type lobby struct {
	// all players in the lobby
	players map[string]*lib.Player

	// channels used to communicate to the main lobby
	say   chan string
	join  chan *lib.Player
	leave chan *lib.Player
}

func NewLobby() *lobby {
	return &lobby{
		players: map[string]*lib.Player{},
		say:     make(chan string),
		join:    make(chan *lib.Player),
		leave:   make(chan *lib.Player),
	}
}

// this is the main player game loop for the lobby
func (l *lobby) Play(player *lib.Player, db *lib.FileStore) {
	// player automatically joins lobby
	l.join <- player
	defer func() {
		l.leave <- player
	}()

	incoming := map[string]interface{}{}
	for {
		if err := player.Receive(&incoming); err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}

		switch {
		case incoming["type"] == "SAY":
			m := incoming["message"].(string)
			l.say <- fmt.Sprintf("Player %s: %s", player.Name(), m)
		case incoming["type"] == "NAME":
			name := incoming["name"].(string)
			player.SetName(name)
			db.Set(*player)
			db.Save()
		default:
			log.Println("Unknown message type", incoming)
		}
	}
}

// this is the main lobby goroutine: all communication must come in through the channels
func (l *lobby) run() {
	for {
		select {
		case player := <-l.join:
			l.players[player.Id()] = player
			l.broadcast(fmt.Sprintln("Player", player.Name(), "joined"))
		case player := <-l.leave:
			delete(l.players, player.Id())
			l.broadcast(fmt.Sprintln("Player", player.Name(), "left"))
		case m := <-l.say:
			l.broadcast(m)
		}
	}
}

type say struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (l *lobby) send(message string, player *lib.Player) {
	msg := say{Type: "say", Message: message}
	if err := player.Send(msg); err != nil {
		if err != io.EOF {
			log.Println("Error sending to player", player.Id(), "so removing them from lobby")
		}
		delete(l.players, player.Id())
	}
}

func (l *lobby) broadcast(message string) {
	for _, player := range l.players {
		l.send(message, player)
	}
}
