package game

import "fmt"

type Move string

const (
	Rock     Move = "rock"
	Paper    Move = "paper"
	Scissors Move = "scissors"
)

func (m Move) IsValidMove() bool {
	return m == Rock || m == Paper || m == Scissors
}

func ParseMove(s string) (Move, error) {
	move := Move(s)
	if !move.IsValidMove() {
		return "", fmt.Errorf("invalid move: %s", s)
	}
	return move, nil
}

func ResolveRound(p1 Move, p2 Move) string {
	if p1 == p2 {
		return "draw"
	}
	if (p1 == Rock && p2 == Scissors) ||
		(p1 == Paper && p2 == Rock) ||
		(p1 == Scissors && p2 == Paper) {
		return "p1"
	}
	return "p2"
}
