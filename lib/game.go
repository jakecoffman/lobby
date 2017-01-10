package lib

type Game interface {
	Init(string, string)
	ID() string
	Run()
	Send(*PlayerCmd)
	Location() string
}
