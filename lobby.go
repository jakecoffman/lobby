package lobby

import (
	"encoding/json"
	"fmt"
	"github.com/jakecoffman/lobby/lib"
	"io"
	"log"
	"reflect"
)

// lobby singleton
var Lobby *lobby

// directory of games by name
var registry map[string]lib.Game

// directory of running games by short ID
var games map[string]lib.Game

func init() {
	Lobby = NewLobby()
	go Lobby.run()
	registry = map[string]lib.Game{}
	games = map[string]lib.Game{}
}

func Register(game lib.Game, name string) {
	registry[name] = game
}

func (l *lobby) RunCallback(id string, game lib.Game) {
	games[id] = game
}

type SimpleCmd struct {
	Player  *lib.Player
	Message string
}

type lobby struct {
	// all players in the lobby
	players map[string]*lib.Player
	games   map[string]lib.Game

	// commands from player
	join       chan *lib.Player
	disconnect chan *lib.Player
	say        chan *SimpleCmd
	rename     chan *SimpleCmd
}

func NewLobby() *lobby {
	return &lobby{
		players:    map[string]*lib.Player{},
		join:       make(chan *lib.Player),
		disconnect: make(chan *lib.Player),
		say:        make(chan *SimpleCmd),
		rename:     make(chan *SimpleCmd),
	}
}

const (
	SAY int = iota
	NAME
	NEW
	JOIN
	LEAVE
	GAME
)

type PlayerCmd struct {
	Type    int             `json:"type"`
	Command json.RawMessage `json:"cmd"`
}

// This is the main player game loop for the lobby. It just reads messages and dispatches
// to the lobby. The lobby is what sends messages back to the players and changes the player objects.
func (l *lobby) Play(player *lib.Player) {
	var game lib.Game

	l.join <- player
	defer func(g lib.Game) {
		if g != nil {
			g.Disconnect(player)
		}
		l.disconnect <- player
	}(game)

	var ok bool

	// rejoin logic
	if player.GameId != "" {
		log.Println("Player is rejoining")
		game, ok = games[player.GameId]
		if ok {
			game.Rejoin(player)
		}
	}

	incoming := &PlayerCmd{}
	cmd := &SimpleCmd{}

	for {
		if err := player.Receive(incoming); err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}

		if incoming.Type != GAME {
			if err := json.Unmarshal(incoming.Command, cmd); err != nil {
				log.Println(string(incoming.Command), err)
				continue
			}
		}
		cmd.Player = player

		switch incoming.Type {
		case SAY:
			l.say <- cmd
		case NAME:
			l.rename <- cmd
		case NEW:
			if game != nil {
				game.Leave(player)
			}
			// create a new game in a goroutine and join it
			game := reflect.New(reflect.TypeOf(registry[cmd.Message])).Elem().Interface().(lib.Game)
			if game != nil {
				go game.Run(l)
				game.Join(player)
			}
		case JOIN:
			game = games[cmd.Message]
			if game != nil {
				game.Join(player)
			}
		case LEAVE:
			if game != nil {
				game.Leave(player)
				game = nil
			}
		case GAME:
			if game != nil {
				game.Send(&incoming.Command)
			}
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
		case player := <-l.disconnect:
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
