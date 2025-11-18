package main

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/matoous/go-nanoid/v2"
)

type Player struct {
	Id        string
	Name      string
	client    *Client
	eventch   chan *Message
	connected bool
	history   []*Message
}

type Client struct {
	conn *websocket.Conn
	Out  chan *Message
	In   chan *Message
}

func (c *Client) ClientId() string {
	return fmt.Sprintf("%s - %s", c.conn.RemoteAddr().Network(), c.conn.RemoteAddr().String())
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		Out:  make(chan *Message),
		In:   make(chan *Message),
	}
}

func NewPlayer(name string) *Player {
	id, _ := gonanoid.New(10)
	return &Player{
		Id:     id,
		Name:   fmt.Sprintf("%s %s", name, id),
		client: nil,
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
