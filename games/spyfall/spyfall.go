package spyfall

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby/lib"
	"github.com/jakecoffman/lobby/server"
	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2"
	"log"
)

var db *mgo.Database

func Install(router *gin.Engine, db *mgo.Database, registry *lib.InMemoryRegistry) {
	db = db
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
}

type Spyfall struct {
	Id         string
	Code       string
	Players    []*Player
	InProgress bool

	cmds chan *lib.PlayerCmd
}

type Player struct {
	*lib.User
	Ready bool
}

func (s *Spyfall) Init(id, code string) {
	s.Id = id
	s.Code = code
	s.cmds = make(chan *lib.PlayerCmd)
	s.Players = []*Player{}
	log.Println("New game initialized", s)
}

func (s *Spyfall) ID() string {
	return s.Id
}

func (s *Spyfall) Run(registry *lib.InMemoryRegistry) {
	for {
		cmd := <-s.cmds

		switch cmd.Type {
		case "NEW":
			s.Players = append(s.Players, &Player{User: cmd.Player})
		case "JOIN":
			// check if this is a rejoin
			found := false
			for i, p := range s.Players {
				if p.ID == cmd.Player.ID {
					found = true
					s.Players[i].User = cmd.Player
					break
				}
			}
			if !found {
				s.Players = append(s.Players, &Player{User: cmd.Player})
			}
		case "DISCONNECT":

		case "LEAVE":
			for i, p := range s.Players {
				if p.ID == cmd.Player.ID {
					s.Players = append(s.Players[:i], s.Players[i+1:]...)
					break
				}
			}
		case "READY":

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
	You     *you
}

type you struct {
	IsSpy    bool
	Location string `json:"omitempty"`
	Role     string `json:"omitempty"`
}

func (s *Spyfall) update() {
	for _, p := range s.Players {
		_ = p.Send(state{
			Type:    "spyfall",
			Spyfall: s,
			You:     &you{IsSpy: true, Location: "France", Role: "Pants"},
		})
	}
}

func (s *Spyfall) String() string {
	return "Spyfall: " + s.Id
}

func (s *Spyfall) Location() string {
	return "spyfall/" + s.Id
}
