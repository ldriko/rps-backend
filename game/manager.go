package game

import (
	"errors"
	"sync"
	"time"
)

type Manager struct {
	games map[string]*Game
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		games: make(map[string]*Game),
	}
}

func (gm *Manager) CreateGame(id, p1, p2 string) (*Game, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if _, exists := gm.games[id]; exists {
		return nil, errors.New("game ID already exists")
	}

	game := NewGame(id, p1, p2)
	gm.games[id] = game
	return game, nil
}

func (gm *Manager) GetGame(id string) (*Game, bool) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	game, exists := gm.games[id]
	return game, exists
}

func (gm *Manager) RemoveGame(id string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if _, exists := gm.games[id]; !exists {
		return errors.New("game not found")
	}

	delete(gm.games, id)
	return nil
}

func (gm *Manager) CleanupExpiredGames(maxAge time.Duration) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	for id, game := range gm.games {
		if time.Since(game.LastActivity) > maxAge {
			delete(gm.games, id)
		}
	}
}
