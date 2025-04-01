package statemanager

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bluelock-go/shared/auth"
	"github.com/bluelock-go/shared/storage/state/token"
)

// State holds the persistent state information
type State struct {
	LastJobExecutionStartTime time.Time                   `json:"lastJobExecutionStartTime"`
	LastJobExecutionEndTime   time.Time                   `json:"lastJobExecutionEndTime"`
	OngoingJobStartTime       time.Time                   `json:"ongoingJobStartTime"`
	RateLimitResetAt          time.Time                   `json:"rateLimitResetAt"`
	CooldownCompletedAt       time.Time                   `json:"cooldownCompletedAt"`
	TokenStates               map[string]token.TokenState `json:"tokenStates"`
}

// StateManager wraps State with a mutex for concurrency safety
type StateManager struct {
	filePath string
	mu       sync.Mutex
	State    State
}

// NewStateManager initializes StateManager and loads existing state
func NewStateManager(filePath string) (*StateManager, error) {
	sm := &StateManager{
		filePath: filePath,
		mu:       sync.Mutex{},
		State: State{
			TokenStates: make(map[string]token.TokenState),
		},
	}

	// Load state from file if it exists
	err := sm.LoadState()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return sm, nil
}

// LoadState reads the state from a JSON file
func (sm *StateManager) LoadState() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := os.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &sm.State)
}

// Sync ToekenStatus With Latest Auth Credentials
func (sm *StateManager) SyncTokenStatusWithLatestAuthCredentials(credentials []auth.Credential) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Create a map to store the latest token states
	latestTokenStates := make(map[string]token.TokenState)

	// Iterate through the credentials and update the token states
	for _, cred := range credentials {
		// TODO: Create a creadKey and use it as the key in the map.
		tokenID := cred.GetUsername()
		tokenState, exists := sm.State.TokenStates[tokenID]
		if !exists {
			tokenState = token.TokenState{}
		}
		tokenState.UpdateTokenStatus(token.TokenActive, time.Now())
		latestTokenStates[tokenID] = tokenState
	}

	sm.State.TokenStates = latestTokenStates

	return sm.saveState()
}

// saveState writes the state to a JSON file
func (sm *StateManager) saveState() error {
	data, err := json.MarshalIndent(sm.State, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.filePath, data, 0644)
}

func (sm *StateManager) SaveStateWithMutex() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.saveState()
}

// ✅ Update Ongoing Job Start Time
func (sm *StateManager) UpdateOngoingJobStartTime(startTime time.Time) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.State.OngoingJobStartTime = startTime
	return sm.saveState()
}

// ✅ Update Last Job Execution Time (Start & End)
func (sm *StateManager) UpdateLastJobExecutionTime(endTime time.Time) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.State.LastJobExecutionStartTime = sm.State.OngoingJobStartTime
	sm.State.LastJobExecutionEndTime = endTime
	return sm.saveState()
}

func (sm *StateManager) UpdateRateLimitResetTime(resetTime time.Time) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.State.RateLimitResetAt = resetTime

	return sm.saveState()
}

// ✅ Replace Token State
func (sm *StateManager) ReplaceTokenState(tokenID string, newState token.TokenState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.State.TokenStates[tokenID] = newState
	return sm.saveState()
}

func (sm *StateManager) SetTokenStatusToRateLimited(tokenID string) error {
	currentTime := time.Now()
	sm.mu.Lock()
	defer sm.mu.Unlock()

	token, exists := sm.State.TokenStates[tokenID]
	if !exists {
		return fmt.Errorf("token %s not found", tokenID)
	}

	token.SetTokenAsExhausted(currentTime)
	sm.State.TokenStates[tokenID] = token

	return sm.saveState()
}

// ✅ Update Last Token Usage Time
func (sm *StateManager) UpdateTokenUsage(tokenID string, usageTime time.Time) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	token, exists := sm.State.TokenStates[tokenID]
	if !exists {
		return fmt.Errorf("token %s not found", tokenID)
	}

	token.UpdateTokenUsage(usageTime)
	sm.State.TokenStates[tokenID] = token

	return sm.saveState()
}

// ResetUsageMetricsForAllTokens resets the usage metrics for all tokens managed by the StateManager.
// It updates the CooldownCompletedAt timestamp to the provided resumeTime and resets the usage metrics
// for each token in the TokenStates map, marking them as active. After updating the state, it persists
// the changes by saving the state.
//
// Parameters:
//   - resumeTime: The time to set as the CooldownCompletedAt timestamp and the reference point for resetting
//     the usage metrics of all tokens.
//
// Returns:
//   - error: An error if saving the updated state fails, otherwise nil.
func (sm *StateManager) ResetUsageMetricsForAllTokens(resumeTime time.Time) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.State.CooldownCompletedAt = resumeTime

	// Reset all token states to active
	for tokenID, token := range sm.State.TokenStates {
		token.ResetUsageMetrics(resumeTime)
		sm.State.TokenStates[tokenID] = token
	}

	return sm.saveState()
}

func (sm *StateManager) GetLeastUsageToken() (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.State.TokenStates) == 0 {
		return "", fmt.Errorf("no tokens available")
	}

	var leastUsedTokenID string
	var leastUsedCount int

	for tokenID, token := range sm.State.TokenStates {
		if leastUsedTokenID == "" || token.SuccessfulUsageCount < leastUsedCount {
			leastUsedTokenID = tokenID
			leastUsedCount = token.SuccessfulUsageCount
		}
	}

	return leastUsedTokenID, nil
}

func (sm *StateManager) UpdateTokenStatus(tokenID string, status token.TokenStatus) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	token, exists := sm.State.TokenStates[tokenID]
	if !exists {
		return fmt.Errorf("token %s not found", tokenID)
	}

	token.UpdateTokenStatus(status, time.Now())
	sm.State.TokenStates[tokenID] = token

	return sm.saveState()
}

// ✅ Get Current Token Status
func (sm *StateManager) GetTokenStatus(tokenID string) (token.TokenStatus, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	token, exists := sm.State.TokenStates[tokenID]
	if !exists {
		return "", false
	}

	return token.Status, true
}
