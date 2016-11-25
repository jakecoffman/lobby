package lib

import (
	"io"

	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2/bson"
	"log"
)

type Player struct {
	ID         string `bson:"_id"`
	Name       string
	connection *websocket.Conn
}

func NewPlayer() *Player {
	return &Player{ID: bson.NewObjectId().Hex()}
}

func (p Player) GetName() string {
	if p.Name != "" {
		return p.Name
	} else {
		return p.ID[len(p.ID) - 5 : len(p.ID)]
	}
}

func (p *Player) Connect(connection *websocket.Conn) {
	p.connection = connection
}

func (p *Player) Disconnect() {
	p.connection.Close()
}

func (p *Player) Send(v interface{}) error {
	return websocket.JSON.Send(p.connection, v)
}

func (p *Player) Receive(v interface{}) error {
	err := websocket.JSON.Receive(p.connection, v)
	if err != nil {
		if err != io.EOF {
			log.Println(err)
		}
		p.Disconnect()
	}
	return err
}
