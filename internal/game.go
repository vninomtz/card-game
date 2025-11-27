package internal

import (
	"errors"
	"fmt"
	"math/rand"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

/*
Rules:
- 2-10 players
- Every player start with 7 cards
- Rest of the cards are placed in a Draw Pile
- Space for a Discard Pile.
- The top card should be placed in the Discard Pile, an the game begins

- Colors: Red, Yellow, Green, Blue, Black
- Numbers: 0, 1, 2, 3, 4, 5, 6, 7, 8, 9
*/

const (
	StateIdle     State = "idle"
	StateStarted  State = "started"
	StateFinished State = "finished"
)

const (
	MAX_CARDS_BY_PLAYER = 7
)

type Game struct {
	Id         string
	Players    []*Player
	Discard    []*Card
	Hands      map[string][]*Card
	Direction  int
	Playing    string // Player Id
	Events     []string
	Rand       *rand.Rand
	State      State
	Winner     string
	rounds     int
	deck       *Deck
	maxPlayers int
}

type GameState struct {
	Status   string
	Players  []*PlayerState
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

func NewGame(seed int64) *Game {
	id, _ := gonanoid.New(10)

	deck := NewDeck(seed)

	g := &Game{
		Id:         id,
		Players:    []*Player{},
		Direction:  1,
		Rand:       rand.New(rand.NewSource(seed)),
		Hands:      make(map[string][]*Card),
		State:      StateIdle,
		deck:       deck,
		maxPlayers: int(deck.Size() / MAX_CARDS_BY_PLAYER),
	}
	return g
}

func (g *Game) AddPlayer(username string) (string, error) {
	if g.State == StateStarted {
		return "", errors.New("Game already started")
	}
	if len(g.Players) == g.maxPlayers {
		return "", errors.New("The game is full. No more players can join")
	}
	p := NewPlayer(username)
	g.Players = append(g.Players, p)
	return p.Id, nil
}

func (g *Game) GetPlayer(playerId string) (*Player, error) {
	for _, p := range g.Players {
		if p.Id == playerId {
			return p, nil
		}
	}
	return nil, errors.New("Player not found")
}

func (g *Game) GetWinner() string {
	return g.Winner
}

func (g *Game) Start() {
	g.deal()
	card, _ := g.draw()
	g.Discard = append(g.Discard, card)
	g.Playing = g.Players[0].Id
	g.State = StateStarted

	g.log(fmt.Sprintf("Game started"))
	g.log(fmt.Sprintf("Players %d: %v", len(g.Players), g.Players))
	g.log(fmt.Sprintf("Turn of Player %s", g.Playing))
	g.log(fmt.Sprintf("Current Card %v", card))
}

func (g *Game) CurrentCard() *Card {
	if len(g.Discard) == 0 {
		return nil
	}
	return g.Discard[len(g.Discard)-1]
}

func (g *Game) CurrentPlayer() *Player {
	var player *Player
	for _, p := range g.Players {
		if p.Id == g.Playing {
			player = p
			break
		}
	}
	return player
}

func (g *Game) GetPlayerHand(playerId string) []*Card {
	return g.Hands[playerId]
}

func (g *Game) Play(playerId string, cardId string) error {
	if g.Playing != playerId {
		return errors.New("Invalid turn")
	}
	hand := g.Hands[playerId]
	found := -1
	var card *Card
	for i, c := range hand {
		if c.Id == cardId {
			found = i
			card = c
			break
		}
	}
	if found == -1 {
		return errors.New("Card not found in your hand")
	}
	if !match(g.CurrentCard(), card) {
		return errors.New("Invalid card")
	}

	g.Hands[playerId] = append(hand[:found], hand[found+1:]...)
	g.Discard = append(g.Discard, card)

	g.log(fmt.Sprintf("Player %s play Card[%s] %s %d. %d left cards", playerId, card.Id, card.Color, card.Number, len(g.Hands[playerId])))

	g.advanceTurn(1)
	g.rounds++

	if len(g.Hands[playerId]) == 0 {
		g.Winner = playerId
		g.Finish()
	}
	return nil
}

func (g *Game) Finish() {
	g.State = StateFinished
	g.log("Game Over")
}

func (g *Game) DrawCard(playerId string) {
	card, drawed := g.draw()
	if drawed {
		g.Hands[playerId] = append(g.Hands[playerId], card)
		g.log(fmt.Sprintf("Player %s draw card %v", playerId, card))
		g.log(fmt.Sprintf("%d cards left on deck", g.deck.Size()))
	} else {
		if !g.existPossibleWinner() {
			g.log("Deck empty, not possible winner found")
			g.log("Draw State")
			g.Finish()
		}
	}
	g.advanceTurn(1)
	g.rounds++
}

func (g *Game) IsGameOver() bool {
	if g.State == StateFinished {
		return true
	}

	return false
}

// func (g *Game) initDeck() {
// 	var Colors []Color = []Color{ColorRed, ColorYellow, ColorGreen, ColorBlue}
// 	for i := 0; i < MAX_NUMBER; i++ {
// 		for _, c := range Colors {
// 			card := NewCard(RankNumber, c, i)
// 			g.Deck = append(g.Deck, card)
// 		}
// 	}
// 	g.log(fmt.Sprintf("Deck created %d", len(g.Deck)))
// }

// func (g *Game) shuffle() {
// 	g.Rand.Shuffle(len(g.Deck), func(i, j int) { g.Deck[i], g.Deck[j] = g.Deck[j], g.Deck[i] })
// 	g.log("Deck suffled")
// }

// func (g *Game) canDraw() bool {
// 	return len(g.Deck) > 0
// }

func (g *Game) draw() (*Card, bool) {
	// var card *Card
	if g.deck.IsEmpty() {
		return nil, false
	}
	// card, g.Deck = g.Deck[len(g.Deck)-1], g.Deck[:len(g.Deck)-1]
	return g.deck.Pop(), true
}

func (g *Game) deal() {
	for _, p := range g.Players {
		g.Hands[p.Id] = []*Card{}
		for i := 0; i < MAX_CARDS_BY_PLAYER; i++ {
			card, _ := g.draw()
			g.Hands[p.Id] = append(g.Hands[p.Id], card)
		}
	}
}

func (g *Game) advanceTurn(num int) {
	index := 0
	for i, p := range g.Players {
		if p.Id == g.Playing {
			index = i
			break
		}
	}
	next := (index + g.Direction*num + len(g.Players)) % len(g.Players)

	g.Playing = g.Players[next].Id
	g.log(fmt.Sprintf("Advance %d turns", num))
	g.log(fmt.Sprintf("Turn of Player %s", g.Playing))

}

func (g *Game) log(ev string) {
	// fmt.Printf("-> Game[%s]: %s\n", g.Id, ev)
	g.Events = append(g.Events, ev)
	//TODO: Maybe put here the logic to send the events to a channel
}

func (g *Game) PrintState() {
	fmt.Printf("Cards discarded %d\n", len(g.Discard))
	fmt.Printf("Cards on Deck %d\n", g.deck.Size())
	fmt.Printf("Game direction %d\n", g.Direction)
	fmt.Printf("Current card: %v\n", g.CurrentCard())
	fmt.Printf("Current player: %v\n", g.CurrentPlayer())
	fmt.Printf("Game rounds %d\n", g.rounds)

	fmt.Println("Players:")
	for _, p := range g.Players {
		hand := g.GetPlayerHand(p.Id)
		fmt.Printf("  Player %s. Cards: %v\n", p.Id, hand)
	}

}

func (g *Game) GetState() GameState {
	players := []*PlayerState{}
	for _, p := range g.Players {
		cards := g.GetPlayerHand(p.Id)
		st := &PlayerState{
			Id:   p.Id,
			Name: p.Name,
		}
		if cards == nil {
			st.LeftCards = 0
		} else {
			st.LeftCards = len(cards)
		}
		players = append(players, st)
	}
	return GameState{
		Players:  players,
		Status:   string(g.State),
		Turn:     g.Playing,
		PlayCard: g.CurrentCard(),
		Winner:   g.Winner,
	}

}

func (g *Game) existPossibleWinner() bool {
	current := g.CurrentCard()
	for _, p := range g.Players {
		for _, card := range g.Hands[p.Id] {
			if match(card, current) {
				return true
			}
		}
	}
	return false
}

func match(c1, c2 *Card) bool {
	if c1.Color == c2.Color {
		return true
	}
	if c1.Number == c2.Number {
		return true
	}
	return false
}

func FindPlayableCard(deck []*Card, toMatch *Card) (bool, *Card) {
	var toPlay *Card
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
