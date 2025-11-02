package main

import (
	"fmt"
	"log"
	"time"
)

func Simulate(players int) {
	g := NewGame(CreatePlayers(players), time.Now().UnixNano())
	g.Start()

	for !g.IsGameOver() {
		player := g.CurrentPlayer()
		hand := g.GetPlayerHand(player.Id)
		has, card := FindPlayableCard(hand, g.CurrentCard())
		if has {
			err := g.Play(player.Id, card)
			if err != nil {
				log.Printf("Player %s play %v, got error %v\n", player.Id, card, err)
			}
		} else {
			g.DrawCard(player.Id)
		}
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("Winner: %s\n", g.Winner)
	g.PrintState()

}

func FindPlayableCard(deck []Card, toMatch Card) (bool, Card) {
	var toPlay Card
	found := false
	for _, c := range deck {
		if match(c, toMatch) {
			toPlay = c
			found = true
			break
		}
	}
	return found, toPlay
}
func CreatePlayers(size int) []Player {
	res := []Player{}
	for i := 0; i <= size; i++ {
		name := fmt.Sprintf("Player %d", i+1)
		res = append(res, NewPlayer(name))
	}
	return res
}
