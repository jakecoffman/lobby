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

// Registry's job is to keep a list of possible games,
// a list of the running games,
// and manage the short "join codes"
type Registry struct {
	sync.RWMutex
	games  map[string]reflect.Type
	gips   map[string]Game   // code -> game in progress
	lookup map[string]string // id -> code
}

func NewInMemoryRegistry() *Registry {
	return &Registry{
		games:  map[string]reflect.Type{},
		gips:   map[string]Game{},
		lookup: map[string]string{},
	}
}

// Register a game
func (r *Registry) Register(game Game, name string) {
	r.Lock()
	r.games[name] = reflect.TypeOf(game).Elem()
	r.Unlock()
}

// Start an instance of a game
func (r *Registry) Start(name string) (Game, error) {
	r.RLock()
	gameType, ok := r.games[name]
	r.RUnlock()
	if !ok {
		return nil, errors.New("Game not found: " + name)
	} else {
		r.Lock()
		defer r.Unlock()
		game := reflect.New(gameType).Interface().(Game)
		id := bson.NewObjectId().Hex()
		// TODO this should be an in-memory key to the game, recycle when everyone disconnects?
		code := gen(7)
		game.Init(id, code, DB)
		r.gips[code] = game
		r.lookup[id] = code
		log.Println("Starting game", name, code, id)
		go game.Run()
		return game, nil
	}

	// TODO save in mongo the lookup maps
}

// Find looks up a game instance by code or ID
func (r *Registry) Find(codeOrId string) (Game, error) {
	r.RLock()
	var game Game
	code, ok := r.lookup[codeOrId] // try id
	if !ok {
		game, ok = r.gips[codeOrId] // try code
	} else {
		game, ok = r.gips[code] // it was an ID
	}
	r.RUnlock()
	if !ok {
		return nil, errors.New("Game not found:" + codeOrId)
	}
	return game, nil
}

// with size = 7, there are 10^7 (10 million) possibilities
func gen(size int) string {
	// 32 chars
	chars := strings.Split("0123456789", "") // no 0, o, 1, l
	code := ""
	for i := 0; i < size; i++ {
		code += chars[rand.Intn(len(chars))]
	}
	return code
}
