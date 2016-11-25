package lib

import (
	"testing"
)

func TestInsertPlayer(t *testing.T) {
	p := NewPlayer()
	p.Name = "Test"
	if err := InsertPlayer(p); err != nil {
		t.Fatal(err)
	}

}