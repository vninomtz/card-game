package main

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/matoous/go-nanoid/v2"
)

type Player struct {
	Id   string
	Name string
	conn *websocket.Conn
}

func NewPlayer(name string) Player {
	id, _ := gonanoid.New(5)
	return Player{
		Id:   id,
		Name: fmt.Sprintf("%s %s", name, id),
	}
}
