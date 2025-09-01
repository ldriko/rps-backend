package game

import "testing"

func TestNewGame(t *testing.T) {
	t.Run("Create new game", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		if game.ID != "game1" {
			t.Errorf("Expected ID to be 'game1', got '%s'", game.ID)
		}
		if game.P1 != "Alice" || game.P2 != "Bob" {
			t.Errorf("Expected players to be 'Alice' and 'Bob', got '%s' and '%s'", game.P1, game.P2)
		}
		if len(game.Rounds) != 0 {
			t.Errorf("Expected no rounds, got %d", len(game.Rounds))
		}
		if game.CurrentRound != nil {
			t.Errorf("Expected no current round, got %v", game.CurrentRound)
		}
		if game.Winner != "" || game.P1Wins != 0 || game.P2Wins != 0 {
			t.Errorf("Expected no winner, got %v", game.Winner)
		}
		if game.CreatedAt.IsZero() {
			t.Errorf("Expected CreatedAt to be set, got zero value")
		}
		if game.LastActivity.IsZero() {
			t.Errorf("Expected UpdatedAt to be set, got zero value")
		}
	})
}

func TestSetPlayerConnected(t *testing.T) {
	t.Run("Set player 1 connected", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		game.SetPlayerConnected("Alice", true)
		if !game.P1Connected {
			t.Errorf("Expected P1Connected to be true, got false")
		}
		if game.P2Connected {
			t.Errorf("Expected P2Connected to be false, got true")
		}
	})

	t.Run("Set player 2 connected", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		game.SetPlayerConnected("Bob", true)
		if !game.P2Connected {
			t.Errorf("Expected P2Connected to be true, got false")
		}
		if game.P1Connected {
			t.Errorf("Expected P1Connected to be false, got true")
		}
	})

	t.Run("Set both players connected", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		game.SetPlayerConnected("Alice", true)
		game.SetPlayerConnected("Bob", true)
		if !game.P1Connected || !game.P2Connected {
			t.Errorf("Expected both players to be connected")
		}
	})

	t.Run("Set player disconnected", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		game.SetPlayerConnected("Alice", true)
		game.SetPlayerConnected("Alice", false)
		if game.P1Connected {
			t.Errorf("Expected P1Connected to be false, got true")
		}
	})
}

func TestIsActive(t *testing.T) {
	t.Run("Both players disconnected", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		if game.IsActive() {
			t.Errorf("Expected game to be inactive, got active")
		}
	})

	t.Run("One player connected", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		game.SetPlayerConnected("Alice", true)
		if !game.IsActive() {
			t.Errorf("Expected game to be active, got inactive")
		}
	})

	t.Run("Both players connected", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		game.SetPlayerConnected("Alice", true)
		game.SetPlayerConnected("Bob", true)
		if !game.IsActive() {
			t.Errorf("Expected game to be active, got inactive")
		}
	})
}

func TestNewRound(t *testing.T) {
	t.Run("Create new round", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		round, err := game.NewRound()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if round == nil {
			t.Errorf("Expected a new round, got nil")
		}
		if game.CurrentRound == nil {
			t.Errorf("Expected current round, got nil")
		}
		if len(game.Rounds) != 0 {
			t.Errorf("Expected no rounds, got %d", len(game.Rounds))
		}
	})

	t.Run("Create multiple rounds", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		for i := 0; i < MaxRounds; i++ {
			_, err := game.NewRound()
			if err != nil {
				t.Errorf("Expected no error on round %d, got %v", i+1, err)
			}
			game.CurrentRound = nil
			game.Rounds = append(game.Rounds, Round{})
		}
		_, err := game.NewRound()
		if err == nil {
			t.Errorf("Expected error when exceeding max rounds, got nil")
		}
	})

	t.Run("Cannot create new round if game is over", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		game.Winner = "Alice"
		_, err := game.NewRound()
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Cannot create new round if max rounds reached", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		for i := 0; i < MaxRounds; i++ {
			game.Rounds = append(game.Rounds, Round{})
		}
		_, err := game.NewRound()
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestPlayRound(t *testing.T) {
	t.Run("Play a valid round", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		_, err := game.NewRound()
		if err != nil {
			t.Fatalf("Expected no error creating new round, got %v", err)
		}
		err = game.PlayRound(Rock, Scissors)
		if err != nil {
			t.Errorf("Expected no error playing round, got %v", err)
		}
		if len(game.Rounds) != 1 {
			t.Errorf("Expected 1 round played, got %d", len(game.Rounds))
		}
		if game.P1Wins != 1 || game.P2Wins != 0 {
			t.Errorf("Expected P1Wins to be 1 and P2Wins to be 0, got %d and %d", game.P1Wins, game.P2Wins)
		}
		if game.Winner != "" {
			t.Errorf("Expected no winner yet, got %s", game.Winner)
		}
	})

	t.Run("Play multiple rounds until game over", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		for i := 0; i < MaxRounds; i++ {
			_, err := game.NewRound()
			if err != nil {
				t.Fatalf("Expected no error creating new round %d, got %v", i+1, err)
			}
			err = game.PlayRound(Rock, Scissors)
			if err != nil {
				t.Fatalf("Expected no error playing round %d, got %v", i+1, err)
			}
		}
		if game.P1Wins != MaxRounds {
			t.Errorf("Expected P1Wins to be %d, got %d", MaxRounds, game.P1Wins)
		}
		if game.Winner != "" {
			t.Errorf("Expected no winner yet, got %s", game.Winner)
		}
		err := game.PlayRound(Rock, Scissors)
		if err == nil {
			t.Errorf("Expected error playing round after max rounds, got nil")
		}
	})

	t.Run("Play draw round", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		_, err := game.NewRound()
		if err != nil {
			t.Fatalf("Expected no error creating new round, got %v", err)
		}
		err = game.PlayRound(Rock, Rock)
		if err != nil {
			t.Errorf("Expected no error playing round, got %v", err)
		}
		if len(game.Rounds) != 1 {
			t.Errorf("Expected 1 round played, got %d", len(game.Rounds))
		}
		if game.P1Wins != 0 || game.P2Wins != 0 {
			t.Errorf("Expected P1Wins and P2Wins to be 0, got %d and %d", game.P1Wins, game.P2Wins)
		}
		if game.Winner != "" {
			t.Errorf("Expected no winner yet, got %s", game.Winner)
		}
	})

	t.Run("Cannot play round if no current round", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		err := game.PlayRound(Rock, Scissors)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Cannot play round if game is over", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		game.Winner = "Alice"
		err := game.PlayRound(Rock, Scissors)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("P2 wins round", func(t *testing.T) {
		game := NewGame("game1", "Alice", "Bob")
		_, err := game.NewRound()
		if err != nil {
			t.Fatalf("Expected no error creating new round, got %v", err)
		}
		err = game.PlayRound(Rock, Paper)
		if err != nil {
			t.Errorf("Expected no error playing round, got %v", err)
		}
		if game.P2Wins != 1 || game.P1Wins != 0 {
			t.Errorf("Expected P2Wins to be 1 and P1Wins to be 0, got %d and %d", game.P2Wins, game.P1Wins)
		}
	})
}
