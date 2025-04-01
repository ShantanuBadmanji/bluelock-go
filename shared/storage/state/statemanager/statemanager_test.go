package statemanager

import (
	"os"
	"testing"
	"time"

	"github.com/bluelock-go/shared/storage/state/token"
	"github.com/stretchr/testify/assert"
)

func TestNewStateManager(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	_, err := NewStateManager(filePath)
	assert.NoError(t, err, "Failed to create StateManager")
}

func TestLoadState(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, err := NewStateManager(filePath)
	assert.NoError(t, err, "Failed to create StateManager")

	// Check if the state is initialized
	assert.NotNil(t, sm.State, "State should not be nil")
	assert.Empty(t, sm.State.TokenStates, "TokenStates should be empty")
}

func TestUpdateOngoingJobStartTime(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, _ := NewStateManager(filePath)
	startTime := time.Now()
	err := sm.UpdateOngoingJobStartTime(startTime)
	assert.NoError(t, err)

	loadedSm, _ := NewStateManager(filePath)
	assert.Equal(t, startTime.Truncate(time.Nanosecond), loadedSm.State.OngoingJobStartTime.Truncate(time.Nanosecond))
}

func TestUpdateLastJobExecutionTime(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, _ := NewStateManager(filePath)
	endTime := time.Now()
	err := sm.UpdateLastJobExecutionTime(endTime)
	assert.NoError(t, err)

	loadedSm, _ := NewStateManager(filePath)
	assert.Equal(t, endTime.Truncate(time.Nanosecond), loadedSm.State.LastJobExecutionEndTime.Truncate(time.Nanosecond))
}

func TestReplaceTokenState(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, _ := NewStateManager(filePath)
	tokenID := "test-token"
	tokenState := token.TokenState{Status: token.TokenActive}
	err := sm.ReplaceTokenState(tokenID, tokenState)
	assert.NoError(t, err)

	loadedSm, _ := NewStateManager(filePath)
	assert.Equal(t, token.TokenActive, loadedSm.State.TokenStates[tokenID].Status)
}

func TestResetUsageMetricsForAllTokens(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, _ := NewStateManager(filePath)
	resumeTime := time.Now()
	err := sm.ResetUsageMetricsForAllTokens(resumeTime)
	assert.NoError(t, err)

	loadedSm, _ := NewStateManager(filePath)
	assert.Equal(t, resumeTime.Truncate(time.Nanosecond), loadedSm.State.CooldownCompletedAt.Truncate(time.Nanosecond))
}

func TestGetLeastUsageToken(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, _ := NewStateManager(filePath)
	token1ID := "token1"
	token2ID := "token2"
	state1 := token.TokenState{SuccessfulUsageCount: 5}
	state2 := token.TokenState{SuccessfulUsageCount: 10}
	sm.ReplaceTokenState(token1ID, state1)
	sm.ReplaceTokenState(token2ID, state2)

	leastUsed, err := sm.GetLeastUsageToken()
	assert.NoError(t, err)
	assert.Equal(t, token1ID, leastUsed)
}
