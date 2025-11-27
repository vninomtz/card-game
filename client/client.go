package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Card struct {
	Id     string `json:"id"`
	Color  string `json:"color"`
	Rank   string `json:"rank"`
	Number int    `json:"number"`
}

type GameState struct {
	Status   string
	Players  []PlayerState
	Turn     string
	PlayCard *Card
	Hand     []*Card
	Winner   string
}

type PlayerState struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	LeftCards int    `json:"left_cards"`
}

type GameClient struct {
	playerId string
	gameId   string
	event    chan string
	msgch    chan *Message
	state    GameState
	cards    []*Card
}

type Message struct {
	Event    string    `json:"action"`
	CardId   string    `json:"cardId"`
	PlayerId string    `json:"playerId"`
	GameId   string    `json:"gameId"`
	State    GameState `json:"state"`
	Cards    []*Card   `json:"cards"`
}
type ActionRequest struct {
	Action   string `json:"action"`
	CardId   string `json:"cardId"`
	PlayerId string `json:"playerId"`
	GameId   string `json:"gameId"`
}

func NewGame() *GameClient {
	return &GameClient{
		event: make(chan string),
		msgch: make(chan *Message),
		state: GameState{},
	}
}

func (g *GameClient) ConnectSSE() {
	url := fmt.Sprintf("http://localhost:8000/games/%s/players/%s/events", g.gameId, g.playerId)
	listener := NewEventListener(url, g.msgch)
	go g.onEvent()
	err := listener.Listen()
	if err != nil {
		log.Println(err)
	}
}

func (g *GameClient) onEvent() {
	for {
		select {
		case msg := <-g.msgch:
			g.state = msg.State
			g.cards = msg.Cards
			g.event <- msg.Event
		}
	}
}

func (g *GameClient) DrawCard() {
	msg := ActionRequest{
		Action:   "DrawCard",
		CardId:   "",
		PlayerId: g.playerId,
		GameId:   g.gameId,
	}
	g.SendRequest(msg)
}

func (g *GameClient) PlayCard(index int) {
	if index > len(g.cards) || index < 0 {
		log.Println("Invalid index")
		return
	}
	cardId := g.cards[index].Id
	msg := ActionRequest{
		Action:   "PlayCard",
		CardId:   cardId,
		PlayerId: g.playerId,
		GameId:   g.gameId,
	}
	g.SendRequest(msg)
}

func (g *GameClient) CreateGame() {
	data := map[string]string{}
	r, err := http.Post("http://localhost:8000/games", "application/json", nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Error to read body %v\n", err)
		return
	}
	g.gameId = data["gameId"]
	g.playerId = data["playerId"]

	go g.ConnectSSE()
}

func (g *GameClient) JoinGame() {
	url := fmt.Sprintf("http://localhost:8000/games/%s/join", g.gameId)
	r, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()

	data := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Error to read body %v\n", err)
		return
	}
	g.gameId = data["gameId"]
	g.playerId = data["playerId"]

	go g.ConnectSSE()
}

func (g *GameClient) StartGame() {
	url := fmt.Sprintf("http://localhost:8000/games/%s/start", g.gameId)
	r, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Println(err)
		return
	}
	if r.StatusCode > 400 {
		log.Printf("Error to start game: Server response with %s status\n", r.Status)
	}
}

func (g *GameClient) SendRequest(msg ActionRequest) {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error to serialize json")
		return
	}
	body := bytes.NewBuffer(b)
	url := fmt.Sprintf("http://localhost:8000/games/%s/players/%s/play", g.gameId, g.playerId)
	r, err := http.Post(url, "application/json", body)
	if err != nil {
		log.Println(err)
		return
	}

	if r.StatusCode > 300 {
		log.Printf("Response of play endpoint with status %s\n", r.Status)
	}
}
