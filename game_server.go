package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type GameServer struct {
	address    string
	port       int
	upgrader   websocket.Upgrader
	register   chan *Client
	unregister chan *Client
	message    chan *Message
	clients    map[*Client]bool
	rooms      map[string]*GameManager
	manager    *GameManager
}

func NewServer(addr string, port int) *GameServer {
	return &GameServer{
		address:    addr,
		port:       port,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		message:    make(chan *Message),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]*GameManager),
		manager:    NewGameManager(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *GameServer) Address() string {
	return fmt.Sprintf("%s:%d", s.address, s.port)
}

func (s *GameServer) Run() error {
	s.routes()

	go s.manager.Run()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		panic(err)
	}
	log.Printf("Starting Server on %s ...\n", s.Address())
	return http.Serve(listener, nil)
}

func (s *GameServer) routes() {
	http.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method != http.MethodPost {
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
			return
		}
		gameId := s.manager.NewGame()

		playerId, _ := s.manager.Join(gameId, "Player")

		payload := map[string]string{
			"gameId":   gameId,
			"playerId": playerId,
		}
		respondWithJSON(w, http.StatusCreated, payload)
	})

	http.HandleFunc("/games/{gameId}/join", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method != http.MethodPost {
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("gameId")
		playerId, err := s.manager.Join(id, "Player")

		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		payload := map[string]string{
			"gameId":   id,
			"playerId": playerId,
		}
		respondWithJSON(w, http.StatusCreated, payload)
	})

	http.HandleFunc("/games/{gameId}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method != http.MethodGet {
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("gameId")

		gm, err := s.manager.GetGame(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		payload := map[string]interface{}{
			"joined":  len(gm.Players),
			"players": gm.Players,
		}
		respondWithJSON(w, http.StatusOK, payload)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		gameId := r.URL.Query().Get("gameId")
		playerId := r.URL.Query().Get("playerId")

		if gameId == "" || playerId == "" {
			http.Error(w, "Missing params", http.StatusBadRequest)
			return
		}
		ws, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = s.manager.ConnectPlayer(gameId, playerId, ws)
		if err != nil {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}
