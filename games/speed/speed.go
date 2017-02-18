package speed

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby/lib"
	"github.com/jakecoffman/lobby/server"
	"golang.org/x/net/websocket"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Deck []Card
var suits = []string{"h", "d", "s", "c"}

func init() {
	rand.Seed(time.Now().Unix())

	for rank := 1; rank < 14; rank++ {
		for _, suit := range suits {
			Deck = append(Deck, Card{Rank: rank, Suit: suit})
		}
	}
}

func Install(router *gin.Engine, registry *lib.Registry) {
	registry.Register(&Speed{}, "speed")
	router.GET("/speed", func(ctx *gin.Context) {
		websocket.Handler(func(conn *websocket.Conn) {
			user, err := server.WSMiddleware(conn)
			if err != nil {
				log.Println(err)
				return
			}

			user.Run(registry)
		}).ServeHTTP(ctx.Writer, ctx.Request)
	})
}

type Speed struct {
	Id       string `bson:"_id"`
	Code     string
	Players  []*Player
	Watchers []*lib.User `json:",omitempty"`
	Started  bool

	Left    Card
	Right   Card
	LSupply []Card
	RSupply []Card

	cmds chan *lib.PlayerCmd
}

type Player struct {
	User  *lib.User
	Ready bool
	Stop  bool // vote to stop game in progress

	hand   []Card
	supply []Card
}

type Card struct {
	Rank int    // 1 - 13 (serializes to base 16 so it's only 1 digit)
	Suit string // h, d, s, c
}

func (c Card) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`"%x%s"`, c.Rank, c.Suit)
	return []byte(str), nil
}

func (c *Card) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) != 4 {
		return errors.New(fmt.Sprintln("length of card needs to be 4, but was", len(s)))
	}
	i, err := strconv.ParseInt(string(s[1]), 16, 64)
	if err != nil {
		return err
	}
	c.Rank = int(i)
	c.Suit = string(s[2])
	return nil
}

// wraps up for marshalling
type state struct {
	Type  string
	Speed *Speed
	You   *you
}

// marshaller for secret information sent only the the player
type you struct {
	Ready bool
	Stop  bool
	Hand  []Card
}

// Init is called when creating a new game
func (s *Speed) Init(id, code string) {
	s.Id = id
	s.Code = code
	s.cmds = make(chan *lib.PlayerCmd)
	s.Players = []*Player{}

	log.Println("New game initialized", s)
}

func (s *Speed) ID() string {
	return s.Id
}

var countdownDuration = 1 * time.Second

func (s *Speed) Run() {
	var cmd *lib.PlayerCmd
	for {
		select {
		case cmd = <-s.cmds:
		}

		log.Println("Processing", cmd.Type, "from player", cmd.Player.ID)

		switch cmd.Type {
		case "NEW":
			s.Players = append(s.Players, &Player{User: cmd.Player})
		case "JOIN":
			// check if this is a rejoin
			found := false
			for i, p := range s.Players {
				if p.User.ID == cmd.Player.ID {
					found = true
					s.Players[i].User = cmd.Player
					break
				}
			}
			if !found {
				if s.Started {
					cmd.Player.Send(&lib.SimpleMsg{
						Type: "error",
						Msg:  "Can't join game in progress",
					})
					continue // don't send update
				}
				if len(s.Players) > 1 {
					cmd.Player.Send(&lib.SimpleMsg{
						Type: "error",
						Msg:  "Game is full",
					})
					continue // don't send update
				}
				s.Players = append(s.Players, &Player{User: cmd.Player})
			}
		case "DISCONNECT":
			for _, p := range s.Players {
				if p.User.ID == cmd.Player.ID {
					p.Ready = false
					break
				}
			}
		// TODO after a certain amount of time leave
		case "LEAVE":
			for i, p := range s.Players {
				if p.User.ID == cmd.Player.ID {
					s.Players = append(s.Players[:i], s.Players[i+1:]...)
					break
				}
			}
			if len(s.Players) == 0 {
				// shut down game
				return
			}
		case "STOP":
			if !s.Started {
				// ignore STOP when game is not started
				continue
			}
			allStop := true
			for _, p := range s.Players {
				if p.User.ID == cmd.Player.ID {
					p.Stop = !p.Stop
				}
				if !p.Stop {
					allStop = false
				}
			}
			if allStop {
				s.Reset()
			}
		case "READY":
			if s.Started {
				// ignore READY when game is started
				continue
			}
			allReady := true
			for i, p := range s.Players {
				if p.User.ID == cmd.Player.ID {
					s.Players[i].Ready = !s.Players[i].Ready
				}
				if p.Ready == false {
					allReady = false
				}
			}
			s.update()

			if !allReady || len(s.Players) != 2 {
				continue
			}

			// TRIGGER GAME START SEQUENCE
			s.Started = true

			// generator!
			draw := make(chan Card)
			go func() {
				r := rand.Perm(52)
				for i := range r {
					draw <- Deck[i]
				}
				close(draw)
			}()

			// assign cards to the game

			for i := 0; i < 5; i++ {
				s.Players[0].hand = append(s.Players[0].hand, <-draw)
				s.Players[1].hand = append(s.Players[1].hand, <-draw)
				s.LSupply = append(s.LSupply, <-draw)
				s.RSupply = append(s.RSupply, <-draw)
			}
			for i := 0; i < 15; i++ {
				s.Players[0].hand = append(s.Players[0].hand, <-draw)
				s.Players[1].hand = append(s.Players[1].hand, <-draw)
			}
			s.Left = <-draw
			s.Right = <-draw

			s.broadcast(&lib.SimpleMsg{
				Type: "starting",
				Msg:  "Game starts in 3",
			})
			time.Sleep(countdownDuration)
			s.broadcast(&lib.SimpleMsg{
				Type: "starting",
				Msg:  "Game starts in 2",
			})
			time.Sleep(countdownDuration)
			s.broadcast(&lib.SimpleMsg{
				Type: "starting",
				Msg:  "Game starts in 1",
			})
			time.Sleep(countdownDuration)
		case "PLAY":
			if !s.Started {
				continue
			}
			play, err := cmd.SimpleCmd()
			if err != nil {
				log.Println(err)
				cmd.Player.Send(&lib.SimpleMsg{
					Type: "error",
					Msg:  "Invalid message",
				})
				continue
			}

			if len(play) != 2 {
				continue
			}

			parts := strings.Split(play, "")
			a, b := parts[0], parts[1]
			num, err := strconv.Atoi(a)
			if err != nil {
				log.Println(err)
				cmd.Player.Send(&lib.SimpleMsg{
					Type: "error",
					Msg:  "Invalid play",
				})
				continue
			}
			// sending -1 indicates they want to play a replacement pile card
			if num < 0 {
				// replacement

				// check if the other player also wants to replace, otherwise set a flag

				// if both players want to replace, change the left and right card to one
				// from the replacement deck

				// if the replacement deck is empty, and there are no plays in hand, game is over? draw?

				break
			}

			var thisguy *Player
			for i, p := range s.Players {
				if p.User.ID == cmd.Player.ID {
					thisguy = s.Players[i]
					break
				}
			}

			if num > len(thisguy.hand) {
				// error
				continue
			}

			playcard := thisguy.hand[num]
			var target *Card
			if b == "0" {
				target = &s.Left
			} else {
				target = &s.Right
			}

			if (
			// special case: Ace can be played on 2 or King
			playcard.Rank == 1 && (target.Rank == 2 || target.Rank == 13)) || (
			// normal case
			playcard.Rank == target.Rank+1 || playcard.Rank == target.Rank-1) {
				// ok: do the play, check if the game is over

				// replace target card with card in hand

				// replace card in hand with card from supply, if any

				// if hand is empty, game is over, WIN!

				break
			}

			// invalid play
			continue // skip update
		default:
			continue
		}
		s.update()
	}
}

func (s *Speed) Send(cmd *lib.PlayerCmd) {
	s.cmds <- cmd
}

func (s *Speed) Reset() {
	s.Started = false
	for _, p := range s.Players {
		p.Ready = false
		p.Stop = false

	}
}

func (s *Speed) update() {
	wg := sync.WaitGroup{}

	wg.Add(len(s.Players))
	for _, player := range s.Players {
		go func(p *Player, speed *Speed) {
			p.User.Send(state{
				Type:  "spyfall",
				Speed: speed,
				You: &you{
					Ready: p.Ready,
					Stop:  p.Stop,
					Hand:  p.hand,
				},
			})
			wg.Done()
		}(player, s)
	}
	wg.Wait()
}

func (s *Speed) broadcast(v interface{}) {
	for _, p := range s.Players {
		_ = p.User.Send(v)
	}
}

func (s *Speed) String() string {
	return "Speed: " + s.Id
}

func (s *Speed) Location() string {
	return "speed/" + s.Id
}
