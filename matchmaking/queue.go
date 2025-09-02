package matchmaking

import (
	"ldriko/rps-backend/game/models"
	"log"
	"sync"
	"time"
)

type QueuedPlayer struct {
	Player   *models.Player
	JoinedAt time.Time
}

type MatchmakingQueue struct {
	players map[string]*QueuedPlayer
	mu      sync.RWMutex
}

func NewQueue() *MatchmakingQueue {
	return &MatchmakingQueue{
		players: make(map[string]*QueuedPlayer),
	}
}

func (q *MatchmakingQueue) AddPlayer(player *models.Player) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.players[player.ID] = &QueuedPlayer{
		Player:   player,
		JoinedAt: time.Now(),
	}
}

func (q *MatchmakingQueue) RemovePlayer(playerID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.players, playerID)
}

func (q *MatchmakingQueue) TryMatch() (*models.Player, *models.Player, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.players) < 2 {
		log.Printf("not enough players to match")
		return nil, nil, false
	}

	var player1, player2 *QueuedPlayer
	for _, p := range q.players {
		if player1 == nil {
			player1 = p
		} else {
			player2 = p
			break
		}
	}

	log.Printf("matched players %s and %s", player1.Player.ID, player2.Player.ID)

	delete(q.players, player1.Player.ID)
	delete(q.players, player2.Player.ID)

	return player1.Player, player2.Player, true
}

func (q *MatchmakingQueue) GetQueueSize() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.players)
}

func (q *MatchmakingQueue) CleanupTimeoutQueuePlayers(maxWait time.Duration) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	for id, qp := range q.players {
		if now.Sub(qp.JoinedAt) > maxWait {
			log.Printf("removing inactive player %s from queue", id)
			delete(q.players, id)
		}
	}
}
