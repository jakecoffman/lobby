package lib

import "encoding/json"

type Lobbyer interface {
	RunCallback(string, Game)
}

type Game interface {
	Run(Lobbyer)
	Join(player *Player)
	Rejoin(player *Player)
	Disconnect(player *Player)
	Leave(player *Player)
	Send(*json.RawMessage)
}
