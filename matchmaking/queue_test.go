package matchmaking

import (
	"fmt"
	"ldriko/rps-backend/game/models"
	"testing"
	"time"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue()
	if q == nil {
		t.Fatalf("Expected queue instance, got nil")
	}
	if len(q.players) != 0 {
		t.Fatalf("Expected 0 players in queue, got %d", len(q.players))
	}
}

func TestAddPlayer(t *testing.T) {
	q := NewQueue()
	player := &models.Player{ID: "player1", Username: "Alice"}
	q.AddPlayer(player)

	if _, exists := q.players[player.ID]; !exists {
		t.Fatalf("Expected player %s to be in queue", player.ID)
	}
	if len(q.players) != 1 {
		t.Fatalf("Expected 1 player in queue, got %d", len(q.players))
	}
}

func TestRemovePlayer(t *testing.T) {
	q := NewQueue()
	player := &models.Player{ID: "player1", Username: "Alice"}
	q.AddPlayer(player)
	q.RemovePlayer(player.ID)

	if _, exists := q.players[player.ID]; exists {
		t.Fatalf("Expected player %s to be removed from queue", player.ID)
	}
	if len(q.players) != 0 {
		t.Fatalf("Expected 0 players in queue, got %d", len(q.players))
	}
}

func TestTryMatch(t *testing.T) {
	t.Run("Not enough players to match", func(t *testing.T) {
		q := NewQueue()
		player := &models.Player{ID: "player1", Username: "Alice"}
		q.AddPlayer(player)

		p1, p2, matched := q.TryMatch()
		if matched {
			t.Fatal("Expected no match, but got one")
		}
		if p1 != nil || p2 != nil {
			t.Fatal("Expected nil players, got non-nil")
		}
	})

	t.Run("Successful match", func(t *testing.T) {
		q := NewQueue()
		player1 := &models.Player{ID: "player1", Username: "Alice"}
		player2 := &models.Player{ID: "player2", Username: "Bob"}
		q.AddPlayer(player1)
		q.AddPlayer(player2)

		p1, p2, matched := q.TryMatch()
		if !matched {
			t.Fatal("Expected a match, but got none")
		}
		if p1 == nil || p2 == nil {
			t.Fatal("Expected non-nil players, got nil")
		}
		if (p1.ID != player1.ID && p1.ID != player2.ID) || (p2.ID != player1.ID && p2.ID != player2.ID) {
			t.Fatal("Matched players do not match the added players")
		}
		if p1.ID == p2.ID {
			t.Fatal("Matched players should be different")
		}
	})

	t.Run("Match with no players", func(t *testing.T) {
		q := NewQueue()

		p1, p2, matched := q.TryMatch()
		if matched {
			t.Fatal("Expected no match, but got one")
		}
		if p1 != nil || p2 != nil {
			t.Fatal("Expected nil players, got non-nil")
		}
	})

	t.Run("Match with only one player", func(t *testing.T) {
		q := NewQueue()
		player := &models.Player{ID: "player1", Username: "Alice"}
		q.AddPlayer(player)

		p1, p2, matched := q.TryMatch()
		if matched {
			t.Fatal("Expected no match, but got one")
		}
		if p1 != nil || p2 != nil {
			t.Fatal("Expected nil players, got non-nil")
		}
	})

	t.Run("Match with more than two players", func(t *testing.T) {
		q := NewQueue()
		player1 := &models.Player{ID: "player1", Username: "Alice"}
		player2 := &models.Player{ID: "player2", Username: "Bob"}
		player3 := &models.Player{ID: "player3", Username: "Charlie"}
		q.AddPlayer(player1)
		q.AddPlayer(player2)
		q.AddPlayer(player3)

		p1, p2, matched := q.TryMatch()
		if !matched {
			t.Fatal("Expected a match, but got none")
		}
		if p1 == nil || p2 == nil {
			t.Fatal("Expected non-nil players, got nil")
		}
		if (p1.ID != player1.ID && p1.ID != player2.ID && p1.ID != player3.ID) ||
			(p2.ID != player1.ID && p2.ID != player2.ID && p2.ID != player3.ID) {
			t.Fatal("Matched players do not match the added players")
		}
		if p1.ID == p2.ID {
			t.Fatal("Matched players should be different")
		}
		if len(q.players) != 1 {
			t.Fatalf("Expected 1 player left in queue, got %d", len(q.players))
		}
	})

	t.Run("Match after removing a player", func(t *testing.T) {
		q := NewQueue()
		player1 := &models.Player{ID: "player1", Username: "Alice"}
		player2 := &models.Player{ID: "player2", Username: "Bob"}
		q.AddPlayer(player1)
		q.AddPlayer(player2)
		q.RemovePlayer(player1.ID)

		p1, p2, matched := q.TryMatch()
		if matched {
			t.Fatal("Expected no match, but got one")
		}
		if p1 != nil || p2 != nil {
			t.Fatal("Expected nil players, got non-nil")
		}
	})

	t.Run("Match after adding and removing players", func(t *testing.T) {
		q := NewQueue()
		player1 := &models.Player{ID: "player1", Username: "Alice"}
		player2 := &models.Player{ID: "player2", Username: "Bob"}
		player3 := &models.Player{ID: "player3", Username: "Charlie"}
		q.AddPlayer(player1)
		q.AddPlayer(player2)
		q.RemovePlayer(player1.ID)
		q.AddPlayer(player3)

		p1, p2, matched := q.TryMatch()
		if !matched {
			t.Fatal("Expected a match, but got none")
		}
		if p1 == nil || p2 == nil {
			t.Fatal("Expected non-nil players, got nil")
		}
		if (p1.ID != player2.ID && p1.ID != player3.ID) || (p2.ID != player2.ID && p2.ID != player3.ID) {
			t.Fatal("Matched players do not match the added players")
		}
		if p1.ID == p2.ID {
			t.Fatal("Matched players should be different")
		}
		if len(q.players) != 0 {
			t.Fatalf("Expected 0 players left in queue, got %d", len(q.players))
		}
	})
}

func TestConcurrentAddRemove(t *testing.T) {
	q := NewQueue()
	playerCount := 100
	done := make(chan bool)

	// Concurrently add players
	for i := 0; i < playerCount; i++ {
		go func(i int) {
			player := &models.Player{ID: "player" + fmt.Sprint(i), Username: "Player" + fmt.Sprint(i)}
			q.AddPlayer(player)
			done <- true
		}(i)
	}

	// Wait for all adds to complete
	for i := 0; i < playerCount; i++ {
		<-done
	}

	if len(q.players) != playerCount {
		t.Fatalf("Expected %d players in queue after adds, got %d", playerCount, len(q.players))
	}

	// Concurrently remove players
	for i := 0; i < playerCount; i++ {
		go func(i int) {
			q.RemovePlayer("player" + fmt.Sprint(i))
			done <- true
		}(i)
	}

	// Wait for all removes to complete
	for i := 0; i < playerCount; i++ {
		<-done
	}

	if len(q.players) != 0 {
		t.Fatalf("Expected 0 players in queue after removes, got %d", len(q.players))
	}
}

func TestGetQueueSize(t *testing.T) {
	q := NewQueue()
	if size := q.GetQueueSize(); size != 0 {
		t.Fatalf("Expected queue size 0, got %d", size)
	}

	player1 := &models.Player{ID: "player1", Username: "Alice"}
	player2 := &models.Player{ID: "player2", Username: "Bob"}
	q.AddPlayer(player1)
	q.AddPlayer(player2)

	if size := q.GetQueueSize(); size != 2 {
		t.Fatalf("Expected queue size 2, got %d", size)
	}

	q.RemovePlayer(player1.ID)
	if size := q.GetQueueSize(); size != 1 {
		t.Fatalf("Expected queue size 1, got %d", size)
	}

	q.RemovePlayer(player2.ID)
	if size := q.GetQueueSize(); size != 0 {
		t.Fatalf("Expected queue size 0, got %d", size)
	}
}

func TestCleanupTimeoutQueuePlayers(t *testing.T) {
	q := NewQueue()
	player1 := &models.Player{ID: "player1", Username: "Alice"}
	player2 := &models.Player{ID: "player2", Username: "Bob"}
	q.AddPlayer(player1)
	q.AddPlayer(player2)

	// Simulate timeout by adjusting JoinedAt
	q.players[player1.ID].JoinedAt = time.Now().Add(-10 * time.Minute)
	q.players[player2.ID].JoinedAt = time.Now().Add(-5 * time.Minute)

	// Cleanup players who have been in queue for more than 6 minutes
	timeout := 6 * time.Minute

	q.CleanupTimeoutQueuePlayers(timeout)

	if size := q.GetQueueSize(); size != 1 {
		t.Fatalf("Expected queue size 1 after cleanup, got %d", size)
	}
	if _, exists := q.players[player1.ID]; exists {
		t.Fatalf("Expected player1 to be removed after cleanup")
	}
	if _, exists := q.players[player2.ID]; !exists {
		t.Fatalf("Expected player2 to remain in queue after cleanup")
	}
}
