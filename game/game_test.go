package game

import "testing"

func TestResolveRound(t *testing.T) {
	tests := []struct {
		name   string
		p1     Move
		p2     Move
		result string
	}{
		// Draw cases
		{"Rock vs Rock", Rock, Rock, "draw"},
		{"Paper vs Paper", Paper, Paper, "draw"},
		{"Scissors vs Scissors", Scissors, Scissors, "draw"},

		// P1 wins cases
		{"Rock vs Scissors", Rock, Scissors, "p1"},
		{"Paper vs Rock", Paper, Rock, "p1"},
		{"Scissors vs Paper", Scissors, Paper, "p1"},

		// P2 wins cases
		{"Rock vs Paper", Rock, Paper, "p2"},
		{"Paper vs Scissors", Paper, Scissors, "p2"},
		{"Scissors vs Rock", Scissors, Rock, "p2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveRound(tt.p1, tt.p2)
			if result != tt.result {
				t.Errorf("Expected %s, got %s", tt.result, result)
			}
		})
	}
}

func TestMoveIsValid(t *testing.T) {
	validMoves := []Move{Rock, Paper, Scissors}
	invalidMoves := []Move{"lizard", "spock", "", "123", "rockk"}

	for _, move := range validMoves {
		t.Run("Valid move: "+string(move), func(t *testing.T) {
			if !move.IsValidMove() {
				t.Errorf("Expected %s to be valid", move)
			}
		})
	}

	for _, move := range invalidMoves {
		t.Run("Invalid move: "+string(move), func(t *testing.T) {
			if move.IsValidMove() {
				t.Errorf("Expected %s to be invalid", move)
			}
		})
	}
}

func TestParseMove(t *testing.T) {
	tests := []struct {
		input       string
		expected    Move
		expectError bool
	}{
		{"rock", Rock, false},
		{"paper", Paper, false},
		{"scissors", Scissors, false},
		{"lizard", "", true},
		{"spock", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run("Parse move: "+tt.input, func(t *testing.T) {
			move, err := ParseMove(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %s, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error for input %s, got %v", tt.input, err)
				}
				if move != tt.expected {
					t.Errorf("Expected move %s, got %s", tt.expected, move)
				}
			}
		})
	}
}
