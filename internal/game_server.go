package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// "time"
)

type GameServer struct {
	message chan *Message
	rooms   map[string]*GameManager
	manager *GameManager
	cfg     *config
}

func NewServer(cfg *config) *GameServer {
	return &GameServer{
		message: make(chan *Message),
		rooms:   make(map[string]*GameManager),
		manager: NewGameManager(),
		cfg:     cfg,
	}
}

func (s *GameServer) Address() string {
	return s.cfg.Addr()
}

func (s *GameServer) Run() error {
	mux := http.NewServeMux()
	s.routes(mux)

	go s.manager.Run()

	srv := &http.Server{
		Addr:    s.cfg.Addr(),
		Handler: s.middlewareCORS(mux),
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		// IdleTimeout:  15 * time.Second,
	}

	log.Printf("Server is running at %s\n", s.cfg.Addr())
	return srv.ListenAndServe()
}

func (s *GameServer) routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /games", s.HandleCreateGame)
	mux.HandleFunc("POST /games/{gameId}/join", s.HandleJoinGame)
	mux.HandleFunc("POST /games/{gameId}/start", s.HandleStartGame)
	mux.HandleFunc("GET /games/{gameId}/players/{playerId}/events", s.HandlePlayerEvents)
	mux.HandleFunc("POST /games/{gameId}/players/{playerId}/play", s.HandlePlayCard)
}

func (s *GameServer) middlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func (s *GameServer) HandleCreateGame(w http.ResponseWriter, r *http.Request) {
	gameId := s.manager.NewGame()
	playerId, _ := s.manager.Join(gameId, "Player")
	payload := map[string]string{
		"gameId":   gameId,
		"playerId": playerId,
	}
	respondWithJSON(w, http.StatusCreated, payload)
}

func (s *GameServer) HandleJoinGame(w http.ResponseWriter, r *http.Request) {
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
	defer r.Body.Close()

	messange := Message{}
	if err := json.NewDecoder(r.Body).Decode(&messange); err != nil {
		log.Printf("Error to read data: %v\n", err)
	}
	s.manager.ProcessMessage(messange)
	respondWithJSON(w, http.StatusOK, nil)
}

func (s *GameServer) HandleStartGame(w http.ResponseWriter, r *http.Request) {
	gameId := r.PathValue("gameId")

	defer r.Body.Close()
	s.manager.StartGame(gameId)
	respondWithJSON(w, http.StatusOK, nil)
}

func (s *GameServer) HandlePlayerEvents(w http.ResponseWriter, r *http.Request) {
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
	w.WriteHeader(code)
	w.Write(response)
	return nil
}
