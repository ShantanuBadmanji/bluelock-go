package statemanager

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/bluelock-go/shared/storage/state/token"
)

func TestStateConcurrency(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, err := NewStateManager(filePath)
	if err != nil {
		t.Fatalf("Failed to create StateManager: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Start multiple goroutines to update different attributes concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			switch id % 3 {
			case 0:
				_ = sm.UpdateOngoingJobStartTime(time.Now())
			case 1:
				_ = sm.UpdateLastJobExecutionTime(time.Now())
			case 2:
				tokenID := fmt.Sprintf("token_%d", id)
				_ = sm.ReplaceTokenState(tokenID, token.TokenState{})
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Validate the state after concurrent updates
	if len(sm.State.TokenStates) != numGoroutines/3 {
		t.Errorf("Expected %d tokens, got %d", numGoroutines/3, len(sm.State.TokenStates))
	}
}

func TestConcurrentTokenUsageUpdates(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, err := NewStateManager(filePath)
	if err != nil {
		t.Fatalf("Failed to create StateManager: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	tokenID := "shared_token"
	tokenState := token.TokenState{Status: token.TokenActive}
	sm.ReplaceTokenState(tokenID, tokenState)

	// Start multiple goroutines to update the same token's usage
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := sm.UpdateTokenUsage(tokenID, time.Now())
			if err != nil {
				t.Errorf("Failed to update token usage: %v", err)
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Validate the state after concurrent updates
	token, exists := sm.State.TokenStates[tokenID]
	if !exists {
		t.Errorf("Token %s not found", tokenID)
	}
	if token.SuccessfulUsageCount != numGoroutines {
		t.Errorf("Expected %d usage count, got %d", numGoroutines, token.SuccessfulUsageCount)
	}
}
