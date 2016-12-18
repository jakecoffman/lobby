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
			user, err := server.Connect(conn)
			if err != nil {
				log.Println(err)
				return
			}

			var game *Spyfall

			defer func() {
				user.Disconnect()
				if game != nil {
					game.cmds <- &lib.PlayerCmd{Type: lib.DISCONNECT, Player: user}
				}
			}()

			// Process incoming messages with a channel since it blocks
			user.Run(registry)
		}).ServeHTTP(ctx.Writer, ctx.Request)
	})
}

type Spyfall struct {
	Id         string
	Code       string
	Players    []*lib.User
	InProgress bool

	cmds chan *lib.PlayerCmd
}

const (
	START int = iota + 100
	STOP
)

func (s *Spyfall) Init(id, code string) {
	s.Id = id
	s.Code = code
	s.cmds = make(chan *lib.PlayerCmd)
	s.Players = []*lib.User{}
	log.Println("New game initialized", s)
}

func (s *Spyfall) ID() string {
	return s.Id
}

func (s *Spyfall) Run(registry *lib.InMemoryRegistry) {
	for {
		cmd := <-s.cmds

		switch cmd.Type {
		case lib.NEW:
			s.Players = append(s.Players, cmd.Player)
		case lib.JOIN:
			log.Println("Player JOIN", cmd.Player.ID)
			// check if this is a rejoin
			for _, p := range s.Players {
				if p.ID == cmd.Player.ID {
					// ok
					break
				}
			}
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
	}
}

func (s *Spyfall) String() string {
	return "Spyfall: " + s.Id
}

func (s *Spyfall) Location() string {
	return "spyfall/" + s.Id
}
