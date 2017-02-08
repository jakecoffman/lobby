package lib

import (
	"encoding/json"
	"github.com/jakecoffman/gorest"
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"sync"
)

const (
	USER = "users"
)

type User struct {
	sync.RWMutex `bson:"-",json:"-"`

	ID   string `bson:"_id"`
	Name string

	// These are for during games when a player does stuff.
	registry   *Registry
	connection Connector
	game       Game
	recv       chan *PlayerCmd
	done       chan struct{}

	Connected bool
}

func (u *User) New() gorest.Resource {
	return &User{}
}

func (u *User) NewList() interface{} {
	return &[]User{}
}

func (u *User) Id(id string) {
	u.ID = id
}

func (u *User) Valid() bool {
	return string(u.ID) != ""
}

func (u User) String() string {
	return "Player: " + u.ID
}

func (u *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        string
		Name      string
		Connected bool
	}{
		ID:        u.ID,
		Name:      u.GetName(),
		Connected: u.Connected,
	})
}

func (u User) GetName() string {
	if u.Name != "" {
		return u.Name
	} else {
		return "Guest " + u.ID[len(u.ID)-5:]
	}
}

func (p *User) Connect(connection Connector) {
	p.Lock()
	p.connection = connection
	p.Connected = true
	p.recv = make(chan *PlayerCmd)
	p.done = make(chan struct{})
	p.Unlock()
}

func (u *User) Disconnect() {
	u.Lock()
	u.Connected = false
	u.connection.Close()
	u.Unlock()
}

func (u *User) Send(v interface{}) error {
	return u.connection.Send(v)
}

func (u *User) receive(v interface{}) error {
	return u.connection.Recv(v)
}

func (p *User) Run(registry *Registry) {
	defer func() {
		p.Disconnect()

		if p.game != nil {
			p.game.Send(&PlayerCmd{Type: "DISCONNECT", Player: p})
		}
	}()

	p.registry = registry

	// Process incoming messages with a channel since it blocks
	go p.sendLoop()

	cmd := &PlayerCmd{}
	for {
		select {
		// Server or the receive channel has signalled a disconnect
		case <-p.done:
			return
		// Handle messages from connection
		case cmd = <-p.recv:
			cmd.Player = p
			p.handle(cmd)
		}
	}
}

// handle handles commands that are common for all games
func (p *User) handle(cmd *PlayerCmd) {
	var err error
	var simple string

	switch cmd.Type {
	case "RENAME":
		if simple, err = cmd.SimpleCmd(); err != nil {
			p.Send(&SimpleMsg{Type: "error", Msg: err.Error()})
			log.Println(err)
			return
		}
		p.Name = simple
	case "NEW":
		if simple, err = cmd.SimpleCmd(); err != nil {
			p.Send(&SimpleMsg{Type: "error", Msg: err.Error()})
			log.Println(err)
			return
		}
		p.game, err = p.registry.Start(simple)
		if err != nil {
			p.Send(&SimpleMsg{Type: "error", Msg: err.Error()})
			log.Println(err)
			return
		}
		log.Println("Player", p.GetName(), "started game", p.game)
	case "FIND": // for finding game & game ID by code or game id
		if simple, err = cmd.SimpleCmd(); err != nil {
			p.Send(&SimpleMsg{Type: "error", Msg: err.Error()})
			log.Println(err)
			return
		}
		var game Game
		game, err = p.registry.Find(simple)
		if err != nil {
			p.Send(&SimpleMsg{Type: "error", Msg: err.Error()})
			log.Println(err)
			return
		}
		p.Send(struct{ Type, Goto string }{"goto", game.Location()})
	case "JOIN": // for joining (by ID or code)
		if simple, err = cmd.SimpleCmd(); err != nil {
			p.Send(&SimpleMsg{Type: "error", Msg: err.Error()})
			log.Println(err)
			return
		}
		p.game, err = p.registry.Find(simple)
		if err != nil {
			p.Send(&SimpleMsg{Type: "error", Msg: err.Error()})
			log.Println(err)
		}
	}
	// send to game if there is one
	if p.game != nil {
		p.game.Send(cmd)
	}
}

func (p *User) sendLoop() {
	var err error
	incoming := &PlayerCmd{}
	for {
		if err = p.receive(incoming); err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			close(p.done)
			p.Disconnect()
			return
		}
		select {
		case p.recv <- incoming:
		case <-p.done:
			return
		}
	}
}

func NewUser() *User {
	return &User{ID: bson.NewObjectId().Hex()}
}

func FindUser(id string) (*User, error) {
	user := &User{}
	err := DB.C(USER).FindId(id).One(user)
	return user, err
}

func InsertUser(user *User) error {
	return DB.C(USER).Insert(user)
}
