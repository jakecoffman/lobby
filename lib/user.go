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

	// These can only be changed from REST API
	ID   string `bson:"_id"`
	Name string

	// These are for during games when a player does stuff.
	registry   Registry
	connection Connector
	game       Game
	Recv       chan *PlayerCmd `bson:"-",json:"-"`
	Done       chan struct{}   `bson:"-",json:"-"`

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

func (u User) GetName() string {
	if u.Name != "" {
		return u.Name
	} else {
		return u.ID[len(u.ID)-5:]
	}
}

func (p *User) Connect(connection Connector, registry Registry) {
	p.Lock()
	p.registry = registry
	p.connection = connection
	p.Connected = true
	p.Recv = make(chan *PlayerCmd)
	p.Done = make(chan struct{})
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

func (p *User) Run(registry Registry) {
	defer func() {
		p.Disconnect()

		if p.game != nil {
			p.game.Send(&PlayerCmd{Type: DISCONNECT, Player: p})
		}
	}()

	p.registry = registry

	// Process incoming messages with a channel since it blocks
	go p.sendLoop()

	// todo automatically rejoin a game disconnected from, or join lobby

	cmd := &PlayerCmd{}
	for {
		select {
		// Server or the receive channel has signalled a disconnect
		case <-p.Done:
			return
		// Handle messages from connection
		case cmd = <-p.Recv:
			cmd.Player = p
			p.handle(cmd)
		}
	}
}

func (p *User) handle(cmd *PlayerCmd) {
	var err error
	var simple string
	log.Println(cmd.Type, string(cmd.Cmd))
	switch cmd.Type {
	case RENAME:
		if simple, err = cmd.SimpleCmd(); err != nil {
			// TODO: send error message
			log.Println(err)
			return
		}
		p.Name = simple
	case NEW:
		if simple, err = cmd.SimpleCmd(); err != nil {
			// TODO: send error message
			log.Println(err)
			return
		}
		p.game, err = p.registry.Start(simple)
		if err != nil {
			// TODO: send error message
			log.Println(err)
			return
		}
		log.Println("Player", p.GetName(), "started game", p.game)
	case FIND: // for finding game & game ID by code or game id
		if simple, err = cmd.SimpleCmd(); err != nil {
			// TODO: send error message
			log.Println(err)
			return
		}
		var game Game
		game, err = p.registry.Find(simple)
		if err != nil {
			log.Println(err)
			// TODO: tell player game was not found
			return
		}
		p.Send(struct{ Type, Goto string }{"goto", game.Location()})
	case JOIN: // for joining (by ID or code)
		if simple, err = cmd.SimpleCmd(); err != nil {
			// TODO: send error message
			log.Println(err)
			return
		}
		p.game, err = p.registry.Find(simple)
		if err != nil {
			log.Println(err)
		}
	case LEAVE:
		// set game to default game?
		p.game.Send(cmd)
		p.game, _ = p.registry.Find("lobby")
		p.game.Send(&PlayerCmd{Type: JOIN, Player: p})
		return
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
			close(p.Done)
			p.Disconnect()
			return
		}
		select {
		case p.Recv <- incoming:
		case <-p.Done:
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

func (u *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        string
		Name      string
		Connected bool
	}{
		u.ID,
		u.Name,
		u.Connected,
	})
}
