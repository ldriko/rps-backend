package main

import "ldriko/rps-backend/game"

func main() {
	result, _ := game.ResolveRound(game.Rock, game.Scissors)

	switch result {
	case "p1":
		println("Player 1 wins!")
	case "p2":
		println("Player 2 wins!")
	default:
		println("It's a draw!")
	}
}
