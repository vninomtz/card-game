package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	// "time"

	"github.com/vninomtz/card-game/internal"
)

func main() {
	games := flag.Int("games", 1, "Games to simulate")
	players := flag.Int("players", 2, "Players for each game")

	flag.Parse()

	manager := internal.NewGameManager()

	var group sync.WaitGroup
	for i := range *games {
		group.Add(1)
		go func(num int) {
			defer group.Done()
			Simulate(num, manager, *players)
		}(i)
	}

	group.Wait()
}

type Player struct {
	Id string
}

type PlayersMap map[string]Player

func Simulate(index int, man *internal.GameManager, size int) {
	gameId := man.NewGame()
	log.Printf("New game simulation %s\n", gameId)
	players := PlayersMap{}

	for i := range size {
		playerid, err := man.Join(gameId, fmt.Sprintf("Player %d", i+1))
		if err != nil {
			log.Printf("Game[%s]: Error to add player to simulation. %s\n", gameId, err)
			continue
		}
		log.Printf("Game[%s]: Player %d added to simulation ", gameId, playerid)
		players[playerid] = Player{Id: playerid}
	}

	man.StartGame(gameId)

	//
	for !man.IsGameOver(gameId) {
		player, hand, _ := man.GetPlayingUser(gameId)
		pCard, _ := man.GetPlayingCard(gameId)
		has, card := internal.FindPlayableCard(hand, pCard)
		if has {
			man.ProcessAction("PlayCard", gameId, player.Id, card.Id)
		} else {
			man.ProcessAction("DrawCard", gameId, player.Id, "")
		}
	}

	winner, _ := man.GetGameWinner(gameId)
	fmt.Printf("Game[%s]: Winner player %s\n", gameId, winner)
}
