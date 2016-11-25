package lib

import (
	"gopkg.in/mgo.v2"
	"log"
)

var Session *mgo.Session
var DB *mgo.Database
var Players *mgo.Collection

func init() {
	Session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		log.Fatal(err)
	}
	DB = Session.DB("lobby")
	Players = DB.C("players")
}

func FindPlayer(id string) (*Player, error) {
	p := &Player{}
	err := Players.FindId(id).One(p)
	return p, err
}

func InsertPlayer(p *Player) error {
	return Players.Insert(p)
}

func UpdatePlayer(p *Player) error {
	_, err := Players.UpsertId(p.ID, p)
	return err
}
