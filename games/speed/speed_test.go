package speed

import (
	"github.com/dustin/gojson"
	"testing"
)

func TestSpeed_CardSerialization(t *testing.T) {
	b, err := json.Marshal(Deck[51])
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"dc"` { // d == 13 == "king", c == "club"
		t.Fatal(string(b))
	}

	var card Card
	if err := json.Unmarshal(b, &card); err != nil {
		t.Fatal(err)
	}
	if card.Suit != "c" || card.Rank != 13 {
		t.Fatal(card)
	}
}
