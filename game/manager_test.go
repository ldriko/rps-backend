package game

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("Expected manager instance, got nil")
	}
	if len(m.games) != 0 {
		t.Fatal("Expected 0 games in manager, got some")
	}
}

func TestCreateGame(t *testing.T) {
	t.Run("Create a new game", func(t *testing.T) {
		m := NewManager()
		game, err := m.CreateGame("Alice", "Bob")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if game == nil {
			t.Fatal("Expected game instance, got nil")
		}
		if game.ID == "" {
			t.Errorf("Expected game to have a valid ID, got '%s'", game.ID)
		}
		if game.P1 != "Alice" || game.P2 != "Bob" {
			t.Errorf("Expected players 'Alice' and 'Bob', got '%s' and '%s'", game.P1, game.P2)
		}
		if len(m.games) != 1 {
			t.Errorf("Expected 1 game in manager, got %d", len(m.games))
		}
	})

	t.Run("Create game with duplicate ID", func(t *testing.T) {
		mockIDs := []string{"fixed-id", "fixed-id", "unique-id"}
		mockGenerator := NewMockUUIDGenerator(mockIDs)
		m := NewManagerWithUUIDGenerator(mockGenerator)

		game1, err := m.CreateGame("Alice", "Bob")
		if err != nil {
			t.Fatalf("Expected no error for first game, got %v", err)
		}
		if game1.ID != "fixed-id" {
			t.Errorf("Expected first game ID to be 'fixed-id', got '%s'", game1.ID)
		}

		game2, err := m.CreateGame("Charlie", "Dave")
		if err != nil {
			t.Fatalf("Expected no error for second game, got %v", err)
		}
		if game2.ID != "unique-id" {
			t.Errorf("Expected second game ID to be 'unique-id', got '%s'", game2.ID)
		}

		if len(m.games) != 2 {
			t.Errorf("Expected 2 games in manager, got %d", len(m.games))
		}
	})

	t.Run("Create game exceeding max retries", func(t *testing.T) {
		mockIDs := []string{"dup-id", "dup-id", "dup-id", "dup-id", "dup-id", "dup-id"}
		mockGenerator := NewMockUUIDGenerator(mockIDs)
		m := NewManagerWithUUIDGenerator(mockGenerator)

		_, err := m.CreateGame("Alice", "Bob")
		if err != nil {
			t.Fatal("Expected no error for first game, got nil")
		}
		_, err = m.CreateGame("Charlie", "Dave")
		if err == nil {
			t.Fatal("Expected error for exceeding max retries, got nil")
		}
	})
}

func TestGetGame(t *testing.T) {
	m := NewManager()
	game, err := m.CreateGame("Alice", "Bob")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	gameID := game.ID
	game.ID = "game1"
	m.games["game1"] = game
	delete(m.games, gameID)

	t.Run("Get existing game", func(t *testing.T) {
		game, exists := m.GetGame("game1")
		if !exists {
			t.Fatal("Expected game instance, got false")
		}
		if game == nil {
			t.Fatal("Expected game instance, got nil")
		}
		if game.ID != "game1" {
			t.Errorf("Expected game ID 'game1', got '%s'", game.ID)
		}
	})

	t.Run("Get non-existing game", func(t *testing.T) {
		game, exists := m.GetGame("nonexistent")
		if exists {
			t.Fatal("Expected game to not exist, got true")
		}
		if game != nil {
			t.Fatal("Expected nil game, got instance")
		}
	})
}

func TestRemoveGame(t *testing.T) {
	m := NewManager()
	game, err := m.CreateGame("Alice", "Bob")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	gameID := game.ID
	game.ID = "game1"
	m.games["game1"] = game
	delete(m.games, gameID)

	t.Run("Remove existing game", func(t *testing.T) {
		err := m.RemoveGame("game1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		_, exists := m.GetGame("game1")
		if exists {
			t.Fatal("Expected game to not exist after removal, but it does")
		}
	})

	t.Run("Remove non-existing game", func(t *testing.T) {
		err := m.RemoveGame("nonexistent")
		if err == nil {
			t.Fatal("Expected error for removing nonexistent game, got nil")
		}
	})
}

func TestCleanupExpiredGames(t *testing.T) {
	m := NewManager()
	game1, err := m.CreateGame("Alice", "Bob")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	game2, err := m.CreateGame("Charlie", "Dave")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	game1ID := game1.ID
	game2ID := game2.ID
	game1.ID = "game1"
	game2.ID = "game2"
	m.games["game1"] = game1
	m.games["game2"] = game2
	delete(m.games, game1ID)
	delete(m.games, game2ID)

	game1.LastActivity = time.Now().Add(-2 * time.Hour)
	game2.LastActivity = time.Now()

	t.Run("Cleanup expired games", func(t *testing.T) {
		m.CleanupExpiredGames(1 * time.Hour)
		if len(m.games) != 1 {
			t.Errorf("Expected 1 game in manager after cleanup, got %d", len(m.games))
		}
		_, exists := m.GetGame("game1")
		if exists {
			t.Fatal("Expected game1 to be removed, but it exists")
		}
		_, exists = m.GetGame("game2")
		if !exists {
			t.Fatal("Expected game2 to exist, but it was removed")
		}
	})
}
