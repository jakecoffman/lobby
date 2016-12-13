package lib

import (
	"gopkg.in/mgo.v2/bson"
)

type Player struct {
	ID         string `bson:"_id"`
	Name       string
	connection Connector
}

func NewPlayer() *Player {
	return &Player{ID: bson.NewObjectId().Hex()}
}

func (p Player) GetName() string {
	if p.Name != "" {
		return p.Name
	} else {
		return p.ID[len(p.ID)-5 : len(p.ID)]
	}
}

func (p *Player) Connect(connection Connector) {
	p.connection = connection
}

func (p *Player) Disconnect() {
	p.connection.Close()
}

func (p *Player) Send(v interface{}) error {
	return p.connection.Send(v)
}

func (p *Player) Receive(v interface{}) error {
	return p.connection.Recv(v)
}
