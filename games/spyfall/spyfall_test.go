package spyfall

import (
	"github.com/jakecoffman/lobby"
	"testing"
)

func TestSpyfall(t *testing.T) {
	lobby.Register(&Spyfall{}, "spyfall")
}
