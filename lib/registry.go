package lib

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"log"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Registry interface {
	Register(Game, string)
	Start(string) (Game, error)
	Find(string) (Game, error)
}

type InMemoryRegistry struct {
	sync.RWMutex
	games  map[string]reflect.Type
	gips   map[string]Game   // code -> game
	lookup map[string]string // id -> code
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		games:  map[string]reflect.Type{},
		gips:   map[string]Game{},
		lookup: map[string]string{},
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
		code := gen(7)
		game.Init(id, code)
		r.gips[code] = game
		r.lookup[id] = code
		log.Println("Starting game", name, code, id)
		go game.Run()
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

// with size = 7, there are 4,294,967,296 possibilities
func gen(size int) string {
	// 32 chars
	chars := strings.Split("23456789abcdefghijkmnpqrstuvwxyz", "") // no 0, o, 1, l
	code := ""
	for i := 0; i < size; i++ {
		code += chars[rand.Intn(len(chars))]
	}
	return code
}
