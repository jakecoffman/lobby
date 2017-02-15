package lib

import "gopkg.in/mgo.v2"

type Game interface {
	Init(string, string, *mgo.Database)
	ID() string
	Run()
	Send(*PlayerCmd)
	Location() string
}
