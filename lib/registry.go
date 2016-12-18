package lib

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"sync"
)

type Registry interface {
	Register(Game, string)
	Start(string) (Game, error)
	Find(string) (Game, error)
}

type InMemoryRegistry struct {
	sync.RWMutex
	games   map[string]reflect.Type
	gips    map[string]Game   // code -> game
	lookup  map[string]string // id -> code
	players map[string]Game   // player -> game
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		games:   map[string]reflect.Type{},
		gips:    map[string]Game{},
		lookup:  map[string]string{},
		players: map[string]Game{},
	}
}

func (r *InMemoryRegistry) Register(game Game, name string) {
	r.Lock()
	r.games[name] = reflect.TypeOf(game).Elem()
	r.Unlock()
}

func (r *InMemoryRegistry) Singleton(game Game, name string) {
	r.Lock()
	r.gips[name] = game
	r.Unlock()
}

func (r *InMemoryRegistry) Start(name string) (Game, error) {
	r.RLock()
	gameType, ok := r.games[name]
	r.RUnlock()
	if !ok {
		return nil, errors.New("Game not found: " + name)
	} else {
		r.Lock()
		defer r.Unlock()
		game := reflect.New(gameType).Interface().(Game)
		// TODO: create 7 digit code users can join each-others games off of
		id := bson.NewObjectId().Hex()
		code := "1"
		game.Init(id, code)
		r.gips[code] = game
		r.lookup[id] = code
		go game.Run(r)
		return game, nil
	}
}

func (r *InMemoryRegistry) Find(codeOrId string) (Game, error) {
	r.RLock()
	var game Game
	code, ok := r.lookup[codeOrId]
	if !ok {
		game, ok = r.gips[codeOrId]
	} else {
		game, ok = r.gips[code]
	}
	r.RUnlock()
	if !ok {
		return nil, errors.New("Game not found:" + codeOrId)
	}
	return game, nil
}

func gen() {

}
