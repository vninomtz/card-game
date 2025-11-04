package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	gameId := flag.String("game", "", "Game Id")
	playerId := flag.String("player", "", "Player Id")
	flag.Parse()

	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:8000/ws?gameId=%s&playerId=%s", *gameId, *playerId), nil)
	if err != nil {
		log.Fatal("Error conectando:", err)
	}
	defer conn.Close()

	msg := map[string]string{}

	msg["Type"] = "Message"
	msg["Action"] = "Test"

	err = conn.WriteJSON(msg)
	if err != nil {
		log.Println("Error al enviar:", err)
		return
	}

	log.Println("Sucess connection")
}
