package lib

type Game interface {
	Init(string, string)
	ID() string
	Run(*InMemoryRegistry)
	Send(*PlayerCmd)
	Location() string
}
