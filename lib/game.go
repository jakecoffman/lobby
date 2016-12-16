package lib

type Game interface {
	Init()
	Run(*InMemoryRegistry)
	Send(*PlayerCmd)
	Location() string
}
