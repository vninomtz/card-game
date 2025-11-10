package main

import (
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Event struct {
	Type     string
	GameId   string
	PlayerId string
	Message  *Message
}

type Message struct {
	Action   string     `json:"action"`
	CardId   string     `json:"cardId"`
	PlayerId string     `json:"playerId"`
	GameId   string     `json:"gameId"`
	State    *GameState `json:"state"`
}
type GameState struct {
	Turn     string
	PlayCard *Card
	Hand     []*Card
}

type GameManager struct {
	games      map[string]*Game
	register   chan *Player
	unregister chan *Player
	event      chan *Event
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*Game),
		event: make(chan *Event),
	}
}

func (r *GameManager) Run() {
	log.Printf("GameManager: Listening for events")
	for {
		select {
		case ev := <-r.event:
			log.Printf("%v\n", ev)
			if ev.Type == "PlayerConnected" {
				r.onPlayerConnect(ev)
			}
		}
	}
}
func (r *GameManager) onPlayerConnect(ev *Event) {
	game, _ := r.games[ev.GameId]
	for _, p := range game.Players {
		if p.client != nil {
			p.client.Out <- &Message{Action: ev.Type}
		}
	}
}

func (r *GameManager) Join(gameId string, username string) (string, error) {
	gm, ok := r.games[gameId]
	if !ok {
		return "", errors.New("Game not found")
	}
	player := NewPlayer(username)
	gm.AddPlayer(player)
	return player.Id, nil
}

func (r *GameManager) NewGame() string {
	gm := NewGame(time.Now().Unix())
	r.games[gm.Id] = gm
	return gm.Id
}

func (r *GameManager) GetGame(id string) (*Game, error) {
	gm, ok := r.games[id]
	if !ok {
		return nil, errors.New("Game not found")
	}
	return gm, nil
}

func (r *GameManager) ProcessMessage(msg Message) {

	switch msg.Action {
	case "StartGame":
		r.StartGame(msg)
	case "PlayCard":
		r.PlayCard(msg)
	case "DrawCard":
		r.DrawCard(msg)
	default:
		log.Println("Unknown Action")
	}
}

func (r *GameManager) PlayCard(msg Message) {
	gm, err := r.GetGame(msg.GameId)
	if err != nil {
		log.Println(err)
		return
	}
	gm.Play(msg.PlayerId, msg.CardId)
	r.SendGameState(gm)
}
func (r *GameManager) DrawCard(msg Message) {
	gm, err := r.GetGame(msg.GameId)
	if err != nil {
		log.Println(err)
		return
	}
	gm.DrawCard(msg.PlayerId)
	r.SendGameState(gm)
}

func (r *GameManager) StartGame(msg Message) {
	gm, err := r.GetGame(msg.GameId)
	if err != nil {
		log.Println(err)
		return
	}
	gm.PrintState()
	gm.Start()

	r.SendGameState(gm)
}

func (r *GameManager) SendGameState(gm *Game) {
	for _, p := range gm.Players {
		state := &GameState{}
		state.Turn = gm.CurrentPlayer().Id
		state.PlayCard = gm.CurrentCard()
		state.Hand = gm.GetPlayerHand(p.Id)
		p.client.Out <- &Message{
			Action:   "GameUpdate",
			GameId:   gm.Id,
			PlayerId: p.Id,
			State:    state,
		}
	}
}

func (r *GameManager) ConnectPlayer(gameId, playerId string, conn *websocket.Conn) error {
	gm, err := r.GetGame(gameId)
	if err != nil {
		return err
	}
	player, err := gm.GetPlayer(playerId)
	if err != nil {
		return err
	}
	player.client = NewClient(conn)

	log.Printf("Connecting %s...\n", playerId)

	r.event <- &Event{Type: "PlayerConnected", PlayerId: playerId, GameId: gameId}
	go r.reader(gameId, player)
	go r.writer(gameId, player)

	return nil
}

func (r *GameManager) writer(gameId string, p *Player) {
	defer func() {
		p.client.conn.Close()
		r.event <- &Event{Type: "PlayerDisconnected", PlayerId: p.Id, GameId: gameId}
	}()

	for {
		select {
		case msg := <-p.client.Out:
			err := p.client.conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Writer: error writing to Player %s, client %s\n", p.Id, p.client.ClientId())
				return
			}
		}
	}

}

func (r *GameManager) reader(gameId string, p *Player) {
	defer func() {
		p.client.conn.Close()
		r.event <- &Event{Type: "PlayerDisconnected", PlayerId: p.Id, GameId: gameId}
	}()

	for {
		var msg *Message
		err := p.client.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error to read json: %v\n", err)
			break
		}
		r.event <- &Event{
			Type:     "Movement",
			PlayerId: p.Id,
			GameId:   gameId,
			Message:  msg,
		}
	}
}
