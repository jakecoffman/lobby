package lib

import (
	"encoding/json"
)

const (
	SAY int = iota
	RENAME
	NEW
	FIND
	JOIN
	LEAVE
	CONNECT
	DISCONNECT
)

type PlayerCmd struct {
	Type int
	Cmd  json.RawMessage

	Player *User  `json:"-"`
	simple string `json:"-"`
}

func (p *PlayerCmd) SimpleCmd() (string, error) {
	if p.simple != "" {
		return p.simple, nil
	}
	simple := &SimpleCmd{}
	if err := json.Unmarshal(p.Cmd, &simple); err != nil {
		return "", err
	}
	p.simple = simple.Msg
	return simple.Msg, nil
}

type SimpleCmd struct {
	Msg string
}
