package internal

import (
	"fmt"
	"testing"
)

func TestDeck(t *testing.T) {
	deck := NewDeck(10)

	if deck.Size() != UNO_DECK_SIZE {
		t.Fatalf("Expected Deck size %d, got %d instead", UNO_DECK_SIZE, deck.Size())
	}
	fmt.Printf("Deck size %d\n", deck.Size())
}
