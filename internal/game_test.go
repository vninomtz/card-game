package internal

import (
	"fmt"
	"testing"
)

func TestGame(t *testing.T) {
	g := NewGame(10)
	g.AddPlayer("Test 1")
	g.AddPlayer("Test 2")
	fmt.Printf("Max number of players %d\n", g.maxPlayers)
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
		err := g.Play(player.Id, card.Id)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		g.DrawCard(player.Id)
	}
	g.PrintState()
}

func TestJoinPlayers(t *testing.T) {
	g := NewGame(10)
	for i := range g.maxPlayers {
		_, err := g.AddPlayer(fmt.Sprintf("Player %d", i+1))
		if err != nil {
			t.Fatalf("Error exception not expected, got %s", err)
		}
	}

	_, err := g.AddPlayer("Not allowed")

	if err == nil {
		t.Fatalf("Expected exception, got nil instead")
	}
	fmt.Printf("Exception to add player: %s\n", err)
	g.Start()

	_, err = g.AddPlayer("Not allowed")
	if err == nil {
		t.Fatalf("Expected exception, got nil instead")
	}

	fmt.Printf("Exception to add player: %s\n", err)

}
