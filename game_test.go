package main

import (
	"fmt"
	"testing"
)

func CreatePlayers(size int) []Player {
	res := []Player{}
	for i := 0; i <= size; i++ {
		name := fmt.Sprintf("Player %d", i+1)
		res = append(res, NewPlayer(name))
	}
	return res
}

func TestInitDeck(t *testing.T) {
	game := NewGame(CreatePlayers(4), 10)
	game.initDeck()
	game.shuffle()

	// for i, c := range game.Deck {
	// 	fmt.Printf("card %d: %s with num %d\n", i+1, c.Color, c.Number)
	// }
	fmt.Printf("Game %s\n", game.Id)
	for k, p := range game.Hands {

		fmt.Printf("Player %s with %d cards\n", k, len(p))
		for _, c := range p {
			fmt.Printf("  %v: \n", c)
		}
	}
}
