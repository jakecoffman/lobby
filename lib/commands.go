package lib

const (
	SAY int = iota
	RENAME
	NEW
	JOIN
	LEAVE
	CONNECT
	DISCONNECT
)

type SimpleCmd struct {
	Player  *Player
	Message string
}
