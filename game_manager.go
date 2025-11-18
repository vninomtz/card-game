package main

import (
	"errors"
	"log"
	"time"
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
	Players  int
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
	for {
		select {
		case ev := <-r.event:
			log.Printf("Event: %s on Game %s executed by Player %s", ev.Type, ev.GameId, ev.PlayerId)
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
	// TODO: Send event
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
	case "PlayCard":
		r.playCard(msg.GameId, msg.PlayerId, msg.CardId)
	case "DrawCard":
		r.drawCard(msg.GameId, msg.PlayerId)
	default:
		log.Println("Unknown Action")
	}
}

func (r *GameManager) playCard(gameId, playerId, cardId string) {
	gm, err := r.GetGame(gameId)
	if err != nil {
		log.Println(err)
		return
	}
	gm.Play(playerId, cardId)
	r.SendGameState(gm)
}
func (r *GameManager) drawCard(gameId, playerId string) {
	gm, err := r.GetGame(gameId)
	if err != nil {
		log.Println(err)
		return
	}
	gm.DrawCard(playerId)
	r.SendGameState(gm)
}

func (r *GameManager) StartGame(gameId string) {
	gm, err := r.GetGame(gameId)
	if err != nil {
		log.Println(err)
		return
	}
	gm.Start()
	// TODO: Send event

	msg := Message{
		Action: "GameStarted",
		GameId: gameId,
	}
	for _, p := range gm.Players {
		if p.connected {
			p.Send(&msg)
		}
	}
	r.SendGameState(gm)
}

func (r *GameManager) SendGameState(gm *Game) {
	for _, p := range gm.Players {
		if p.connected {
			state := &GameState{}
			state.Players = len(gm.Players)
			state.Turn = gm.CurrentPlayer().Id
			state.PlayCard = gm.CurrentCard()
			state.Hand = gm.GetPlayerHand(p.Id)

			p.eventch <- &Message{
				Action:   "GameUpdated",
				GameId:   gm.Id,
				PlayerId: p.Id,
				State:    state,
			}
		}
	}
}

func (r *GameManager) PlayerConnected(gameId string, player *Player) error {
	gm, err := r.GetGame(gameId)
	if err != nil {
		return err
	}
	msg := Message{
		Action:   "PlayerConnected",
		GameId:   gameId,
		PlayerId: player.Id,
		State: &GameState{
			Players: len(gm.Players),
		},
	}
	for _, p := range gm.Players {
		if p.connected {
			p.Send(&msg)
		}
	}
	return nil
}

func (r *GameManager) FindUser(gameId, playerId string) (*Player, error) {
	gm, err := r.GetGame(gameId)
	if err != nil {
		return nil, err
	}
	player, err := gm.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	return player, nil
}
