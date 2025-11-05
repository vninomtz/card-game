package main

import (
	"testing"
)

func TestInitDeck(t *testing.T) {
	g := NewGame(10)
	g.Start()

	if g.CurrentCard().Id == "" {
		t.Fatalf("Expected card to init Game, got %v instead", g.CurrentCard())
	}

	player := g.CurrentPlayer()
	if player.Id == "" {
		t.Fatalf("Expected a player, got %v instead", player)
	}
	hand := g.GetPlayerHand(player.Id)
	if len(hand) != 7 {
		t.Fatalf("Expected start with 7 cards, got %d instead", len(hand))
	}
	has, card := FindPlayableCard(hand, g.CurrentCard())
	if has {
		err := g.Play(player.Id, card)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		g.DrawCard(player.Id)
	}
}
