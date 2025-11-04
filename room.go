package main

import gonanoid "github.com/matoous/go-nanoid/v2"

type Room struct {
	code    string
	started bool
	players map[string]*Player
}

func NewRoom() *Room {
	id, _ := gonanoid.New(10)
	return &Room{
		code:    id,
		started: false,
		players: make(map[string]*Player),
	}
}

func (r *Room) Join(player *Player) {
	r.players[player.Id] = player
}
