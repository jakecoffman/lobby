package lib

import (
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"sync"
)

type Player struct {
	sync.RWMutex

	recv       chan *PlayerCmd
	done       chan struct{}

	ID         string `bson:"_id"`
	Name       string
	Connected  bool
	connection Connector

	registry   Registry
	game       Game
}

func NewPlayer(registry Registry) *Player {
	return &Player{ID: bson.NewObjectId().Hex(), registry: registry}
}

func (p Player) GetName() string {
	if p.Name != "" {
		return p.Name
	} else {
		return p.ID[len(p.ID) - 5 : len(p.ID)]
	}
}

func (p *Player) Connect(connection Connector) {
	p.Lock()
	p.connection = connection
	p.Connected = true
	p.Unlock()
}

func (p *Player) Run(registry Registry) {
	var err error
	p.game, err = registry.Find("lobby")
	if err != nil {
		log.Println(err)
		return
	}
	p.game.Send(&PlayerCmd{Type: JOIN, Player: p})

	defer func() {
		p.connection.Close()
		close(p.done)
		close(p.recv)

		if p.game != nil {
			p.game.Send(&PlayerCmd{Type: DISCONNECT})
		}
	}()

	// Process incoming messages with a channel since it blocks
	go p.sendLoop()

	// todo automatically rejoin a game disconnected from, or join lobby

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

func (p *Player) handle(cmd *PlayerCmd) {
	var err error
	var simple string
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
	case JOIN:
		if simple, err = cmd.SimpleCmd(); err != nil {
			// TODO: send error message
			log.Println(err)
			return
		}
		p.game, err = p.registry.Find(simple)
	case LEAVE:
		// set game to default game?
		p.game.Send(cmd)
		p.game, _ = p.registry.Find("lobby")
		p.game.Send(&PlayerCmd{Type: JOIN, Player: p})
		return
	}
	// send to game if there is one
	p.game.Send(cmd)
}

func (p *Player) Send(v interface{}) error {
	return p.connection.Send(v)
}

func (p *Player) receive(v interface{}) error {
	return p.connection.Recv(v)
}

func (p *Player) sendLoop() {
	var err error
	incoming := &PlayerCmd{}
	for {
		if err = p.receive(incoming); err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			close(p.done)
			return
		}
		select {
		case p.recv <- incoming:
		case <-p.done:
			return
		}
	}
}
