package server

import (
	"encoding/json"
	"ldriko/rps-backend/game"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

type Connection struct {
	ws       *websocket.Conn
	send     chan []byte
	playerID string
	gameID   string
	mu       sync.Mutex
}

type Message struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type Server struct {
	gm        *game.Manager
	conns     map[string]*Connection
	gameConns map[string][]*Connection
	mu        sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		gm:        game.NewManager(),
		conns:     make(map[string]*Connection),
		gameConns: make(map[string][]*Connection),
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		http.Error(w, "player_id is required", http.StatusBadRequest)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		http.Error(w, "websocket upgrade failed", http.StatusInternalServerError)
		return
	}

	conn := &Connection{
		ws:       ws,
		send:     make(chan []byte, 256),
		playerID: playerID,
	}

	s.registerConnection(conn)

	go conn.readMessages(s)
	go conn.writeMessages()
}

func (s *Server) registerConnection(conn *Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if oldConn, exists := s.conns[conn.playerID]; exists {
		oldConn.ws.Close()
	}

	s.conns[conn.playerID] = conn
	log.Printf("player %s connected", conn.playerID)
}

func (s *Server) unregisterConnection(conn *Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conns[conn.playerID] == conn {
		delete(s.conns, conn.playerID)
	}

	s.removePlayerFromGame(conn, conn.gameID)

	if conn.gameID != "" {
		if game, exists := s.gm.GetGame(conn.gameID); exists {
			game.SetPlayerConnected(conn.playerID, false)
		}
	}

	log.Printf("player %s disconnected from game", conn.playerID)
}

func (s *Server) addPlayerToGame(conn *Connection, gameID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.gm.GetGame(gameID)
	if !exists {
		log.Printf("game %s not found", gameID)
		return
	}

	conn.gameID = gameID
	s.gameConns[gameID] = append(s.gameConns[gameID], conn)
}

func (s *Server) removePlayerFromGame(conn *Connection, gameID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn.gameID == "" {
		log.Printf("trying to remove from game but conn.gameID is empty")
		return
	}

	if connections, exists := s.gameConns[gameID]; exists {
		for i, c := range connections {
			if c == conn {
				s.gameConns[gameID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}

		if len(s.gameConns[gameID]) == 0 {
			delete(s.gameConns, gameID)
		}
	}

	conn.gameID = ""
}

func (conn *Connection) readMessages(s *Server) {
	defer func() {
		s.unregisterConnection(conn)
		conn.close()
	}()

	for {
		var msg Message
		err := conn.ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		s.handleMessage(conn, &msg)
	}
}

func (conn *Connection) writeMessages() {
	defer conn.ws.Close()

	for message := range conn.send {
		err := conn.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("error writing message: %v", err)
			return
		}
	}

	conn.ws.WriteMessage(websocket.CloseMessage, []byte{})
}

func (conn *Connection) close() {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	close(conn.send)
}

func (conn *Connection) SendMessage(msg Message) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshalling message: %v", err)
		return
	}

	select {
	case conn.send <- data:
	default:
		close(conn.send)
	}
}

func (s *Server) handleMessage(conn *Connection, msg *Message) {
	switch msg.Type {
	case "join_game":
		s.handleJoinGame(conn, msg.Data)
	case "make_move":
		s.handleMakeMove(conn, msg.Data)
	case "start_round":
		s.handleStartRound(conn, msg.Data)
	default:
		log.Printf("unknown message type: %s", msg.Type)
	}
}

func (s *Server) handleJoinGame(conn *Connection, data map[string]any) {
	gameID, ok := data["game_id"].(string)
	if !ok {
		log.Printf("player %s tried to join a non-existent game", conn.playerID)
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": "invalid game_id"},
		})
		return
	}

	game, exists := s.gm.GetGame(gameID)
	if !exists {
		log.Printf("creating new game for player %s", conn.playerID)
		var err error
		game, err = s.gm.CreateGame(conn.playerID, "")
		if err != nil {
			log.Printf("failed to create a game for player %s: %v", conn.playerID, err)
			conn.SendMessage(Message{
				Type: "error",
				Data: map[string]any{"message": err.Error()},
			})
			return
		}
	}

	s.addPlayerToGame(conn, game.ID)
	game.SetPlayerConnected(conn.playerID, true)

	conn.SendMessage(Message{
		Type: "game_joined",
		Data: map[string]any{
			"gameID": game.ID,
			"game":   game,
		},
	})

	s.broadcastToGame(gameID, Message{
		Type: "player_joined",
		Data: map[string]any{"playerID": conn.playerID},
	}, conn.playerID)
}

func (s *Server) handleMakeMove(conn *Connection, data map[string]any) {
	gameID := conn.gameID
	if gameID == "" {
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": "not in a game"},
		})
		return
	}

	moveStr, ok := data["move"].(string)
	if !ok {
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": "invalid move"},
		})
		return
	}

	gm, exists := s.gm.GetGame(gameID)
	if !exists {
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": "game not found"},
		})
		return
	}

	if gm.CurrentRound == nil {
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": "no active round"},
		})
		return
	}

	move, err := game.ParseMove(moveStr)
	if err != nil {
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": "invalid move"},
		})
		return
	}

	switch conn.playerID {
	case gm.P1:
		gm.CurrentRound.P1 = move
	case gm.P2:
		gm.CurrentRound.P2 = move
	default:
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": "not a player in this game"},
		})
		return
	}

	if gm.CurrentRound.P1.IsValidMove() && gm.CurrentRound.P2.IsValidMove() {
		err := gm.PlayRound(gm.CurrentRound.P1, gm.CurrentRound.P2)
		if err != nil {
			conn.SendMessage(Message{
				Type: "error",
				Data: map[string]any{"message": err.Error()},
			})
			return
		}

		s.broadcastToGame(gameID, Message{
			Type: "round_played",
			Data: map[string]any{
				"round": gm.Rounds[len(gm.Rounds)-1],
				"game":  gm,
			},
		}, "")
	}
}

func (s *Server) handleStartRound(conn *Connection, _ map[string]any) {
	gameID := conn.gameID
	if gameID == "" {
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": "not in a game"},
		})
		return
	}

	gm, exists := s.gm.GetGame(gameID)
	if !exists {
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": "game not found"},
		})
		return
	}

	_, err := gm.NewRound()
	if err != nil {
		conn.SendMessage(Message{
			Type: "error",
			Data: map[string]any{"message": err.Error()},
		})
		return
	}

	s.broadcastToGame(gameID, Message{
		Type: "round_started",
		Data: map[string]any{
			"round": gm.CurrentRound,
			"game":  gm,
		},
	}, "")
}

func (s *Server) broadcastToGame(gameID string, msg Message, excludePlayerID string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conns, exists := s.gameConns[gameID]
	if !exists {
		return
	}

	for _, conn := range conns {
		if conn.playerID != excludePlayerID {
			conn.SendMessage(msg)
		}
	}
}

func (s *Server) sendToPlayer(playerID string, msg Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if conn, exists := s.conns[playerID]; exists {
		conn.SendMessage(msg)
	}
}
