package game

import (
	"errors"
	"time"
)

type Move string

const (
	Rock     Move = "rock"
	Paper    Move = "paper"
	Scissors Move = "scissors"
)

type Round struct {
	P1        Move
	P2        Move
	Winner    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m Move) IsValidMove() bool {
	return m == Rock || m == Paper || m == Scissors
}

func ParseMove(s string) (Move, error) {
	move := Move(s)
	if !move.IsValidMove() {
		return "", errors.New("invalid move")
	}
	return move, nil
}

func ResolveRound(p1 Move, p2 Move) (string, error) {
	if !p1.IsValidMove() || !p2.IsValidMove() {
		return "", errors.New("invalid move")
	}

	if p1 == p2 {
		return "draw", nil
	}
	if (p1 == Rock && p2 == Scissors) ||
		(p1 == Paper && p2 == Rock) ||
		(p1 == Scissors && p2 == Paper) {
		return "p1", nil
	}
	return "p2", nil
}
