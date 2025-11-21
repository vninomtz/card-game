package internal

import (
	"fmt"

	"github.com/matoous/go-nanoid/v2"
)

type Player struct {
	Id        string
	Name      string
	eventch   chan *Message
	connected bool
	history   []*Message
}

func NewPlayer(name string) *Player {
	id, _ := gonanoid.New(10)
	return &Player{
		Id:   id,
		Name: fmt.Sprintf("%s %s", name, id),
	}
}

func (p *Player) Connect() {
	p.connected = true
	p.eventch = make(chan *Message)
}
func (p *Player) Disconnect() {
	p.connected = false
	close(p.eventch)
}
func (p *Player) Send(msg *Message) {
	p.history = append(p.history, msg)
	if p.connected {
		p.eventch <- msg
	}
}
