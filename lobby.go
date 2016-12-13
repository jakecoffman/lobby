package lobby

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/jakecoffman/lobby/lib"
)

var Lobby *lobby

func init() {
	Lobby = NewLobby()
	go Lobby.run()
}

type SimpleCmd struct {
	Player  *lib.Player
	Message string
}

type Game interface {
	Join()
	Leave()
	Send()
}

type lobby struct {
	// all players in the lobby
	players map[string]*lib.Player
	games   map[string]Game

	// commands from player
	join   chan *lib.Player
	leave  chan *lib.Player
	say    chan *SimpleCmd
	rename chan *SimpleCmd
}

func NewLobby() *lobby {
	return &lobby{
		players: map[string]*lib.Player{},
		join:    make(chan *lib.Player),
		leave:   make(chan *lib.Player),
		say:     make(chan *SimpleCmd),
		rename:  make(chan *SimpleCmd),
	}
}

const (
	SAY int = iota
	NAME
	NEW
	JOIN
	GAME
)

type PlayerCmd struct {
	Type    int             `json:"type"`
	Command json.RawMessage `json:"cmd"`
}

// This is the main player game loop for the lobby. It just reads messages and dispatches
// to the lobby. The lobby is what sends messages back to the players and changes the player objects.
func (l *lobby) Play(player *lib.Player) {
	// player automatically joins lobby
	l.join <- player
	defer func() {
		l.leave <- player
	}()

	incoming := &PlayerCmd{}
	for {
		if err := player.Receive(incoming); err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}

		cmd := &SimpleCmd{}
		if err := json.Unmarshal(incoming.Command, cmd); err != nil {
			log.Println(string(incoming.Command), err)
			continue
		}
		cmd.Player = player

		switch {
		case incoming.Type == SAY:
			l.say <- cmd
		case incoming.Type == NAME:
			l.rename <- cmd
		case incoming.Type == NEW:
			// create a new game in a goroutine and join it
		case incoming.Type == JOIN:
			// look up by game join number and join go routine
		case incoming.Type == GAME:
			// forward message to the game
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
			l.players[player.ID] = player
			l.broadcast(fmt.Sprint("Player ", player.GetName(), " joined"))
		case player := <-l.leave:
			delete(l.players, player.ID)
			player.Disconnect()
			l.broadcast(fmt.Sprint("Player ", player.GetName(), " left"))
		case cmd := <-l.say:
			l.broadcast(fmt.Sprintf("%s: %s", cmd.Player.GetName(), cmd.Message))
		case cmd := <-l.rename:
			previousName := cmd.Player.GetName()
			cmd.Player.Name = cmd.Message
			if err := lib.UpdatePlayer(cmd.Player); err != nil {
				log.Println(err)
			}
			l.broadcast(fmt.Sprintf("%s is now known as %s", previousName, cmd.Player.GetName()))
		}
	}
}

type say struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (l *lobby) send(message string, player *lib.Player) {
	if err := player.Send(say{Type: "say", Message: message}); err != nil {
		if err != io.EOF {
			log.Println("Error sending to player", player.ID, ":", err)
		}
	}
}

func (l *lobby) broadcast(message string) {
	for _, player := range l.players {
		l.send(message, player)
	}
}
