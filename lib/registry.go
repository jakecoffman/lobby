package lib

import (
	"errors"
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
	gips    map[string]Game // code -> game
	players map[string]Game // player -> game
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		games:   map[string]reflect.Type{},
		gips:    map[string]Game{},
		players: map[string]Game{},
	}
}

func (r *InMemoryRegistry) Register(game Game, name string) {
	r.Lock()
	r.games[name] = reflect.TypeOf(game)
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
		return nil, errors.New("Game not found")
	} else {
		r.Lock()
		defer r.Unlock()
		game := reflect.New(gameType).Elem().Interface().(Game)
		game.Init()
		// TODO: create 7 digit code users can join each-others games off of
		code := "1"
		r.gips[code] = game
		go game.Run(r)
		return game, nil
	}
}

func (r *InMemoryRegistry) Find(code string) (Game, error) {
	r.RLock()
	game, ok := r.gips[code]
	r.RUnlock()
	if !ok {
		return nil, errors.New("Game not found")
	}
	return game, nil
}

func gen() {

}
