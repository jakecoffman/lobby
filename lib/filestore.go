package lib

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type FileStore struct {
	sync.RWMutex
	cookies map[string]Player
}

func NewFileStore() *FileStore {
	return &FileStore{cookies: map[string]Player{}}
}

func (c *FileStore) Get(key string) (Player, bool) {
	c.RLock()
	defer c.RUnlock()
	p, ok := c.cookies[key]
	return p, ok
}

func (c *FileStore) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.cookies, key)
}

func (c *FileStore) Set(player Player) {
	c.Lock()
	c.cookies[player.Id()] = player
	c.Unlock()
	c.Save()
}

func (c *FileStore) Save() {
	c.RLock()
	defer c.RUnlock()
	file, err := os.Create("cookies.json")
	if err != nil {
		log.Println(err)
		return
	}
	if err = json.NewEncoder(file).Encode(c.cookies); err != nil {
		log.Println(err)
	}
}

func (c *FileStore) Load() error {
	c.Lock()
	defer c.Unlock()
	file, err := os.Open("cookies.json")
	if err != nil {
		log.Println(err)
		return err
	}
	if err = json.NewDecoder(file).Decode(&c.cookies); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
