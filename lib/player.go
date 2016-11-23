package lib

import (
	"io"

	"encoding/json"
	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2/bson"
)

type Player struct {
	id         string
	name       string
	connection *websocket.Conn
}

func NewPlayer() *Player {
	return &Player{id: bson.NewObjectId().Hex()}
}

func (p Player) Id() string {
	return p.id
}

func (p Player) Name() string {
	if p.name != "" {
		return p.name
	} else {
		return p.id[len(p.id)-5 : len(p.id)]
	}
}

func (p *Player) SetName(name string) {
	p.name = name
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
	if err == io.EOF {
		p.Disconnect()
	}
	return err
}

type marshalPlayer struct {
	Id   string
	Name string
}

func (p Player) Marshal() ([]byte, error) {
	return json.Marshal(marshalPlayer{
		Id:   p.id,
		Name: p.name,
	})
}

func (p Player) Unmarshal(data []byte) error {
	m := marshalPlayer{}
	err := json.Unmarshal(data, &m)
	p.id = m.Id
	p.name = m.Name
	return err
}
