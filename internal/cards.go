package internal

import (
	"math/rand"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Color string
type Rank string
type State string

const (
	ColorRed    Color = "red"
	ColorYellow Color = "yellow"
	ColorGreen  Color = "green"
	ColorBlue   Color = "blue"
	ColorBlack  Color = "black"
)
const (
	RankNumber Rank = "number"
)

type Card struct {
	Id     string `json:"id"`
	Color  Color  `json:"color"`
	Rank   Rank   `json:"rank"`
	Number int    `json:"number"`
}

func NewCard(r Rank, c Color, n int) *Card {
	id, _ := gonanoid.New(10)
	return &Card{
		Id:     id,
		Color:  c,
		Number: n,
		Rank:   r,
	}
}

type Deck struct {
	items []*Card
	rnd   *rand.Rand
}

func NewDeck(seed int64) *Deck {
	var deck []*Card
	var Colors []Color = []Color{ColorRed, ColorYellow, ColorGreen, ColorBlue}
	for i := 0; i < MAX_NUMBER; i++ {
		for _, c := range Colors {
			card := NewCard(RankNumber, c, i)
			deck = append(deck, card)
		}
	}
	d := &Deck{
		items: deck,
		rnd:   rand.New(rand.NewSource(seed)),
	}
	d.suffle()
	return d
}

func (d *Deck) suffle() {
	d.rnd.Shuffle(len(d.items), func(i, j int) { d.items[i], d.items[j] = d.items[j], d.items[i] })
}

func (d *Deck) IsEmpty() bool {
	return len(d.items) == 0
}

func (d *Deck) Pop() *Card {
	if len(d.items) == 0 {
		return nil
	}
	i := len(d.items) - 1
	card := d.items[i]
	d.items = d.items[:i]
	return card
}

func (d *Deck) Size() int {
	return len(d.items)
}
