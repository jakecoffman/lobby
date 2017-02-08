package spyfall

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby/lib"
	"github.com/jakecoffman/lobby/server"
	"golang.org/x/net/websocket"
	"log"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func Install(router *gin.Engine, registry *lib.Registry) {
	registry.Register(&Spyfall{}, "spyfall")
	router.GET("/spyfall", func(ctx *gin.Context) {
		websocket.Handler(func(conn *websocket.Conn) {
			user, err := server.WSMiddleware(conn)
			if err != nil {
				log.Println(err)
				return
			}

			user.Run(registry)
		}).ServeHTTP(ctx.Writer, ctx.Request)
	})
	router.GET("/spyfall/locations", func(ctx *gin.Context) {
		ctx.Writer.Write(locations())
	})
}

type Spyfall struct {
	Id       string
	Code     string
	Players  []*Player
	Watchers []*lib.User `json:",omitempty"`
	Started  bool

	EndsAt time.Time

	cmds  chan *lib.PlayerCmd
	timer *time.Timer
}

type Player struct {
	User  *lib.User
	Ready bool
	Stop  bool // vote to stop game in progress
	First bool

	spy      bool // secret information sent only to the player
	location string
	role     string
}

// wraps up for marshalling
type state struct {
	Type    string
	Spyfall *Spyfall
	You     *you
}

// marshaller for secret information sent only the the player
type you struct {
	Spy      bool
	Location string `json:",omitempty"`
	Role     string `json:",omitempty"`
	Ready    bool
	Stop     bool
}

// Init is called when creating a new game
func (s *Spyfall) Init(id, code string) {
	s.Id = id
	s.Code = code
	s.cmds = make(chan *lib.PlayerCmd)
	s.Players = []*Player{}
	s.timer = time.NewTimer(1 * time.Minute)
	if !s.timer.Stop() {
		<-s.timer.C
	}
	log.Println("New game initialized", s)
	// TODO: save to mongo
}

func (s *Spyfall) ID() string {
	return s.Id
}

func (s *Spyfall) Run() {
	var cmd *lib.PlayerCmd
	var ok bool
	for {
		select {
		case cmd, ok = <-s.cmds:
			if !ok {
				s.cmds = nil
				return
			}
		case <-s.timer.C:
			log.Println("Time ended")
			s.Reset()
			s.update()
			continue
		}

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
				if !s.timer.Stop() {
					<-s.timer.C
				}
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

			if !allReady {
				break
			}

			// TRIGGER GAME START SEQUENCE
			s.Started = true
			location := placeData.Locations[rand.Intn(len(placeData.Locations))]
			roles := placeData.Roles[location]

			roleCursor := rand.Intn(len(roles))
			s.Players[rand.Intn(len(s.Players))].spy = true
			s.Players[rand.Intn(len(s.Players))].First = true

			for _, p := range s.Players {
				if p.spy {
					continue
				}
				p.role = roles[roleCursor]
				p.location = location
				roleCursor += 1
				if roleCursor > len(roles) {
					roleCursor = 0
				}
			}

			s.broadcast(&lib.SimpleMsg{
				Type: "starting",
				Msg:  "Game starts in 3",
			})
			time.Sleep(time.Second)
			s.broadcast(&lib.SimpleMsg{
				Type: "starting",
				Msg:  "Game starts in 2",
			})
			time.Sleep(time.Second)
			s.broadcast(&lib.SimpleMsg{
				Type: "starting",
				Msg:  "Game starts in 1",
			})
			time.Sleep(time.Second)
			s.EndsAt = time.Now().Add(8 * time.Minute)
			s.timer = time.NewTimer(8 * time.Minute)
		default:
			continue
		}
		s.update()
		// TODO: save to Mongo
	}
}

func (s *Spyfall) Send(cmd *lib.PlayerCmd) {
	s.cmds <- cmd
}

func (s *Spyfall) Reset() {
	s.Started = false
	for _, p := range s.Players {
		p.Ready = false
		p.Stop = false
		p.First = false
		p.spy = false
		p.location = ""
		p.role = ""
	}
}

func (s *Spyfall) update() {
	for _, p := range s.Players {
		_ = p.User.Send(state{
			Type:    "spyfall",
			Spyfall: s,
			You: &you{
				Spy:      p.spy,
				Location: p.location,
				Role:     p.role,
				Ready:    p.Ready,
				Stop:     p.Stop,
			},
		})
	}
}

func (s *Spyfall) broadcast(v interface{}) {
	for _, p := range s.Players {
		_ = p.User.Send(v)
	}
}

func (s *Spyfall) String() string {
	return "Spyfall: " + s.Id
}

func (s *Spyfall) Location() string {
	return "spyfall/" + s.Id
}
