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
	Event    string    `json:"action"`
	CardId   string    `json:"cardId"`
	PlayerId string    `json:"playerId"`
	GameId   string    `json:"gameId"`
	State    GameState `json:"state"`
	Cards    []*Card   `json:"cards"`
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
	defer r.mu.Unlock()
	gm, ok := r.games[gameId]
	if !ok {
		return "", errors.New("Game not found")
	}
	return gm.AddPlayer(username)
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

func (r *GameManager) ProcessAction(action, gameId, playerId, cardId string) {
	switch action {
	case "PlayCard":
		r.playCard(gameId, playerId, cardId)
	case "DrawCard":
		r.drawCard(gameId, playerId)
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

	ev := "CardPlayed"
	if gm.State == StateFinished {
		ev = "GameFinished"
	}
	r.Broadcast(gm, ev)
}
func (r *GameManager) drawCard(gameId, playerId string) {
	gm, err := r.GetGame(gameId)
	if err != nil {
		log.Println(err)
		return
	}
	gm.DrawCard(playerId)

	r.Broadcast(gm, "CardDrawed")
}

func (r *GameManager) StartGame(gameId string) {
	gm, err := r.GetGame(gameId)
	if err != nil {
		log.Println(err)
		return
	}
	gm.Start()

	r.Broadcast(gm, "GameStarted")
}

func (r *GameManager) Broadcast(gm *Game, event string) {
	st := gm.GetState()
	for _, p := range gm.Players {
		if p.connected {
			msg := &Message{
				Event:  event,
				GameId: gm.Id,
				State:  st,
				Cards:  gm.GetPlayerHand(p.Id),
			}
			p.Send(msg)
		}
	}
}

func (r *GameManager) ConnectPlayer(gameId, playerId string) (chan *Message, error) {
	player, err := r.FindUser(gameId, playerId)
	if err != nil {
		return nil, err
	}
	player.Connect()
	log.Printf("Game[%s]: Player %s connected", gameId, playerId)

	go r.PlayerConnected(gameId, player)
	return player.eventch, nil
}

func (r *GameManager) DisconnectPlayer(gameId, playerId string) {
	player, err := r.FindUser(gameId, playerId)
	if err != nil {
		log.Printf("Disconnect error: Player %s not found\n", playerId)
		return
	}

	player.Disconnect()
	log.Printf("Game[%s]: Player %s disconnected", gameId, playerId)
}

func (r *GameManager) PlayerConnected(gameId string, player *Player) error {
	gm, err := r.GetGame(gameId)
	if err != nil {
		return err
	}
	// msg := Message{
	// 	Event:    "PlayerConnected",
	// 	GameId:   gameId,
	// 	PlayerId: player.Id,
	// 	State:    &GameState{
	// 		// Players: len(gm.Players),
	// 	},
	// }
	// for _, p := range gm.Players {
	// 	if p.connected {
	// 		p.Send(&msg)
	// 	}
	// }

	r.Broadcast(gm, "PlayerConnected")
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
