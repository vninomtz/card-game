package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("UNO Game CLI Client")
	client := NewClient()

	client.Start()
}

type GameClient struct {
	playerId string
	gameId   string
	conn     *websocket.Conn
	event    chan string
}
type Message struct {
	Action   string `json:"action"`
	CardId   string `json:"cardId"`
	PlayerId string `json:"playerId"`
	GameId   string `json:"gameId"`
}

func NewClient() *GameClient {
	return &GameClient{
		event: make(chan string),
	}
}

func (g *GameClient) Connect() {
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:8000/ws?gameId=%s&playerId=%s", g.gameId, g.playerId), nil)
	if err != nil {
		log.Fatal("Error to connect with web socket:", err)
	}
	g.conn = conn

	go g.reader()
}

func (g *GameClient) reader() {
	defer func() {
		g.conn.Close()
		g.event <- "Client disconnected"
	}()

	for {
		msg := map[string]string{}
		err := g.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error to read json %v\n", err)
			break
		}
		g.event <- msg["action"]
	}
}

func (g *GameClient) onEvent() {
	for {
		select {
		case event := <-g.event:
			fmt.Printf("Game: %s\n", event)
			fmt.Print(" > ")
		}
	}
}

func (g *GameClient) Start() {
	scanner := bufio.NewScanner(os.Stdin)
	go g.onEvent()
	for {
		fmt.Print(" > ")
		scanner.Scan()
		str := scanner.Text()

		g.ProcessInput(str)
	}
}

func (g *GameClient) ProcessInput(value string) {
	words := strings.Fields(value)
	if len(words) == 0 {
		return
	}

	cmd := words[0]

	switch cmd {
	case "create":
		g.CreateGame()
	case "join":
		if len(words) < 2 {
			log.Println("Expected gameId")
			return
		}
		g.gameId = words[1]
		g.JoinGame()
	case "start":
		g.StartGame()
	default:
		g.event <- "Unkonwn command"
	}
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

	g.Connect()

	g.event <- fmt.Sprintf("Game created with id %s. Player Id %s", g.gameId, g.playerId)
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

	g.Connect()

	g.event <- fmt.Sprintf("Joined to game with id %s. Player Id %s", g.gameId, g.playerId)
}

func (g *GameClient) StartGame() {
	msg := Message{
		Action:   "StartGame",
		CardId:   "",
		PlayerId: g.playerId,
		GameId:   g.gameId,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error to serialize json")
		return
	}

	body := bytes.NewBuffer(b)
	r, err := http.Post("http://localhost:8000/games/play", "application/json", body)

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Play response %s\n", r.Status)
}
