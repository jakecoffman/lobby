package lib

import "encoding/json"

const (
	SAY int = iota
	RENAME
	NEW
	JOIN
	LEAVE
	CONNECT
	DISCONNECT
)

type PlayerCmd struct {
	Type    int             `json:"type"`
	Command json.RawMessage `json:"cmd"`

	Player  *Player `json:"-"`
	simple  string `json:"-"`
}

func (p *PlayerCmd) SimpleCmd() (string, error) {
	if p.simple != "" {
		return p.simple, nil
	}
	simple := &SimpleCmd{}
	if err := json.Unmarshal(p.Command, &simple); err != nil {
		return "", err
	}
	p.simple = simple.Message
	return simple.Message, nil
}

type SimpleCmd struct {
	Message string
}
