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
	http.HandleFunc("/games", s.HandleCreateGame)
	http.HandleFunc("/games/{gameId}/join", s.HandleJoinGame)
	http.HandleFunc("/games/{gameId}/start", s.HandleStartGame)
	http.HandleFunc("/games/{gameId}/players/{playerId}/events", s.HandlePlayerEvents)
	http.HandleFunc("/games/{gameId}/players/{playerId}/play", s.HandlePlayCard)
}

func (s *GameServer) HandleCreateGame(w http.ResponseWriter, r *http.Request) {
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
}

func (s *GameServer) HandleJoinGame(w http.ResponseWriter, r *http.Request) {
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
}

func (s *GameServer) HandlePlayCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	messange := Message{}
	if err := json.NewDecoder(r.Body).Decode(&messange); err != nil {
		log.Printf("Error to read data: %v\n", err)
	}
	s.manager.ProcessMessage(messange)
	respondWithJSON(w, http.StatusOK, nil)
}

func (s *GameServer) HandleStartGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}
	gameId := r.PathValue("gameId")

	defer r.Body.Close()
	s.manager.StartGame(gameId)
	respondWithJSON(w, http.StatusOK, nil)
}

func (s *GameServer) HandlePlayerEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	gameId := r.PathValue("gameId")
	playerId := r.PathValue("playerId")

	player, err := s.manager.FindUser(gameId, playerId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
		return
	}
	notify := r.Context().Done()
	defer func() {
		player.Disconnect()
		log.Printf("Player %s disconnected", player.Id)
	}()

	player.Connect()
	flusher.Flush()
	go s.manager.PlayerConnected(gameId, player)
	for {
		select {
		case <-notify:
			log.Println("Client close the connection")
			return
		case msg := <-player.eventch:
			payload, err := json.Marshal(msg)
			if err != nil {
				log.Println("Error to marshal message: %v", msg)
				break
			}
			fmt.Fprintf(w, "event: %s\n", "Message")
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
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
