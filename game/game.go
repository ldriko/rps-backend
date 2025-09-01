package game

import (
	"errors"
	"sync"
	"time"
)

const MaxRounds = 3

type Game struct {
	ID string
	P1 string
	P2 string

	Rounds       []Round
	CurrentRound *Round

	P1Wins int
	P2Wins int
	Winner string

	CreatedAt    time.Time
	LastActivity time.Time

	P1Connected bool
	P2Connected bool
	mu          sync.Mutex
}

func NewGame(id, p1, p2 string) *Game {
	return &Game{
		ID:           id,
		P1:           p1,
		P2:           p2,
		Rounds:       []Round{},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
}

func (g *Game) SetPlayerConnected(player string, connected bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	switch player {
	case g.P1:
		g.P1Connected = connected
	case g.P2:
		g.P2Connected = connected
	}
	g.LastActivity = time.Now()
}

func (g *Game) IsActive() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.P1Connected || g.P2Connected
}

func (g *Game) NewRound() (*Round, error) {
	if g.Winner != "" {
		return nil, errors.New("game over")
	} else if len(g.Rounds) >= MaxRounds {
		return nil, errors.New("maximum rounds reached")
	}

	newRound := Round{}
	g.CurrentRound = &newRound
	return &newRound, nil
}

func (g *Game) PlayRound(p1Move, p2Move Move) error {
	if g.Winner != "" {
		return errors.New("game over")
	} else if len(g.Rounds) >= MaxRounds {
		return errors.New("maximum rounds reached")
	} else if g.CurrentRound == nil {
		return errors.New("no current round to play")
	}

	result, err := ResolveRound(p1Move, p2Move)
	if err != nil {
		return err
	}

	g.CurrentRound.P1 = p1Move
	g.CurrentRound.P2 = p2Move
	g.CurrentRound.Winner = result

	switch result {
	case "p1":
		g.P1Wins++
	case "p2":
		g.P2Wins++
	}

	g.Rounds = append(g.Rounds, *g.CurrentRound)
	g.CurrentRound = nil

	return nil
}
