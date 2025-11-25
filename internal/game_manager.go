package internal

import (
	"errors"
	"log"
	"sync"
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
	Winner   string
}

type GameManager struct {
	games      map[string]*Game
	register   chan *Player
	unregister chan *Player
	event      chan *Event
	mu         sync.RWMutex
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
	r.mu.Lock()
	gm, ok := r.games[gameId]
	if !ok {
		r.mu.Unlock()
		return "", errors.New("Game not found")
	}
	player := NewPlayer(username)
	gm.AddPlayer(player)
	r.mu.Unlock()
	// TODO: Send event
	return player.Id, nil
}

func (r *GameManager) NewGame() string {
	gm := NewGame(time.Now().Unix())
	r.mu.Lock()
	r.games[gm.Id] = gm
	r.mu.Unlock()
	return gm.Id
}

func (r *GameManager) GetGame(id string) (*Game, error) {
	r.mu.Lock()
	gm, ok := r.games[id]
	r.mu.Unlock()
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

	if gm.State == StateFinished {
		winner := gm.GetWinner()
		msg := Message{
			Action: "GameFinished",
			GameId: gameId,
			State: &GameState{
				Winner: winner,
			},
		}
		r.Broadcast(gm, &msg)
	}

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
	r.Broadcast(gm, &msg)
	r.SendGameState(gm)
}

func (r *GameManager) Broadcast(gm *Game, msg *Message) {
	for _, p := range gm.Players {
		if p.connected {
			p.Send(msg)
		}
	}
}

func (r *GameManager) SendGameState(gm *Game) {
	for _, p := range gm.Players {
		if p.connected {
			state := &GameState{}
			state.Players = len(gm.Players)
			state.Turn = gm.CurrentPlayer().Id
			state.PlayCard = gm.CurrentCard()
			state.Hand = gm.GetPlayerHand(p.Id)
			state.Winner = gm.GetWinner()

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

func (r *GameManager) IsGameOver(gameId string) bool {
	gm, err := r.GetGame(gameId)
	if err != nil {
		return true
	}

	return gm.IsGameOver()
}

func (r *GameManager) GetPlayingUser(gameId string) (*Player, []*Card, error) {
	gm, err := r.GetGame(gameId)
	if err != nil {
		return nil, nil, err
	}
	player := gm.CurrentPlayer()
	hand := gm.GetPlayerHand(player.Id)

	return player, hand, nil
}

func (r *GameManager) GetPlayingCard(gameId string) (*Card, error) {
	gm, err := r.GetGame(gameId)
	if err != nil {
		return nil, err
	}
	hand := gm.CurrentCard()

	return hand, nil
}
func (r *GameManager) GetGameWinner(gameId string) (string, error) {
	gm, err := r.GetGame(gameId)
	if err != nil {
		return "", err
	}

	return gm.Winner, nil
}
