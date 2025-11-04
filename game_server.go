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
	rooms      map[string]*Room
}

type Client struct {
	conn     *websocket.Conn
	playerId string
	gameId   string
}

type Message struct {
	Type     string `json:"type"`
	Action   string `json:"action"`
	PlayerId string
	GameId   string
}

func NewClient(conn *websocket.Conn, playerId, gameId string) *Client {
	return &Client{
		conn:     conn,
		playerId: playerId,
		gameId:   gameId,
	}
}

func (c *Client) ClientId() string {
	return fmt.Sprintf("%s - %s", c.conn.RemoteAddr().Network(), c.conn.RemoteAddr().String())
}

func NewServer(addr string, port int) *GameServer {
	return &GameServer{
		address:    addr,
		port:       port,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		message:    make(chan *Message),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]*Room),
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

	go s.run()
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
		room := NewRoom()
		player := NewPlayer("Player")
		room.Join(&player)
		s.rooms[room.code] = room

		payload := map[string]string{
			"gameId":   room.code,
			"playerId": player.Id,
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
		room, ok := s.rooms[id]
		if !ok {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}
		player := NewPlayer("Player")
		room.Join(&player)
		payload := map[string]string{
			"gameId":   room.code,
			"playerId": player.Id,
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
		room, ok := s.rooms[id]
		if !ok {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"joined":  len(room.players),
			"players": room.players,
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
		_, ok := s.rooms[gameId]
		if !ok {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}
		_, ok = s.rooms[gameId].players[playerId]
		if !ok {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}
		ws, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		client := NewClient(ws, playerId, gameId)
		s.register <- client
	})
}

func (s *GameServer) run() {
	log.Printf("ServerManager: Listening for events")
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
			go s.reader(client)
			log.Printf("ServerManager: Client %s connected\n", client.ClientId())
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				log.Printf("ServerManager: Client %s disconnected, %d connected clients", client.ClientId(), len(s.clients))
			}
		case msg := <-s.message:
			log.Printf("ServerManager: Player %s send message %v", msg.PlayerId, msg)
		}
	}
}
func (s *GameServer) writer(client *Client) {

}
func (s *GameServer) reader(client *Client) {
	defer func() {
		client.conn.Close()
		s.unregister <- client
	}()
	var msg *Message
	for {
		err := client.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error to read json: %v\n", err)
			break
		}
		msg.PlayerId = client.playerId
		msg.GameId = client.gameId
		s.message <- msg
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
