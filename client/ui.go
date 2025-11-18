package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type State int

const (
	StateMainMenu State = iota
	StateCreateGame
	StateJoinGame
	StateLobby
	StateGame
	StateGameFinished
	StateExit
)

func Clear() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}
func Red(s string) string    { return "\033[31m" + s + "\033[0m" }
func Green(s string) string  { return "\033[32m" + s + "\033[0m" }
func Yellow(s string) string { return "\033[33m" + s + "\033[0m" }

type App struct {
	State   State
	game    *GameClient
	inputch chan string
}

func NewApp() *App {
	gm := NewGame()
	return &App{
		State:   StateMainMenu,
		game:    gm,
		inputch: make(chan string),
	}
}

func (a *App) Run() {
	go a.readInput()

	for {
		a.render()

		select {
		case input := <-a.inputch:
			a.handleInput(input)
		case ev := <-a.game.event:
			a.handleEvent(ev)
		}
	}
}

func (a *App) render() {
	Clear()
	switch a.State {
	case StateMainMenu:
		a.MainMenu()
	case StateLobby:
		a.LobbyGame()
	case StateGame:
		a.ShowGame()
	case StateGameFinished:
		a.ShowGameEnd()
	case StateExit:
		fmt.Println("Bye!")
	}
}

func (a *App) readInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		a.inputch <- scanner.Text()
	}
}

func (a *App) handleInput(input string) {
	words := strings.Fields(input)
	if len(words) == 0 {
		return
	}
	cmd := words[0]
	switch cmd {
	case "create":
		a.game.CreateGame()
		a.State = StateLobby
	case "join":
		if len(words) < 2 {
			return
		}
		a.game.gameId = words[1]
		a.game.JoinGame()
		a.State = StateLobby
	case "start":
		a.game.StartGame()
		a.State = StateGame
	case "play":
		if len(words) < 2 {
			fmt.Printf("Missing Card to play")
			return
		}
		index, _ := strconv.Atoi(words[1])
		a.game.PlayCard(index)
	case "draw":
		a.game.DrawCard()
	case "home":
		a.State = StateMainMenu
	case "exit":
		a.State = StateExit
		os.Exit(0)
	}
}

func (a *App) handleEvent(ev string) {
	if ev == "GameStarted" {
		a.State = StateGame
	}
	if ev == "GameFinished" {
		a.State = StateGameFinished
	}
}

func (a *App) MainMenu() {
	fmt.Println("CLI UNO GAME")
	fmt.Println("Menu:")
	fmt.Println("- Create")
	fmt.Println("- Join")
	fmt.Println("- Exit")
	fmt.Println()
	fmt.Print("> ")
}

func (a *App) LobbyGame() {
	fmt.Println("CLI UNO GAME")
	fmt.Printf("Game: %s  ", a.game.gameId)
	fmt.Printf("Players: %d\n", a.game.state.Players)
	fmt.Println("Waiting for more players...")
	fmt.Println()
	fmt.Println("Menu:")
	fmt.Println("- Start")
	fmt.Println("- Exit")
	fmt.Println()
	fmt.Print("> ")
}

func (a *App) ShowGameEnd() {
	fmt.Println("CLI UNO GAME")
	fmt.Printf("Game: %s finished\n", a.game.gameId)
	fmt.Printf("Winner Player %s \n", a.game.state.Winner)
	fmt.Println()
	fmt.Println("Menu:")
	fmt.Println("- Home")
	fmt.Println("- Exit")
	fmt.Println()
	fmt.Print("> ")

}

func (a *App) ShowGame() {
	fmt.Printf("CLI UNO GAME Game: %s   Player: %s\n", a.game.gameId, a.game.playerId)
	fmt.Printf("Players: %d\n", a.game.state.Players)
	fmt.Print("Turn: ")
	if a.game.playerId == a.game.state.Turn {
		fmt.Printf("Your turn\n")
	} else {
		fmt.Printf("Player %s\n", a.game.state.Turn)
	}
	fmt.Printf("Current Card: %d %s", a.game.state.PlayCard.Number, a.game.state.PlayCard.Color)
	fmt.Println()
	fmt.Println("Hand:")
	for i, c := range a.game.state.Hand {
		fmt.Printf("-> Card[%d]: %d %s\n", i, c.Number, c.Color)
	}

	fmt.Println()
	fmt.Println("Menu:")
	fmt.Println("- Play {index of card}")
	fmt.Println("- Draw")
	fmt.Println("- Exit")
	fmt.Println()
	fmt.Print("> ")
}
