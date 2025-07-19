package bitbucketcloud

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/auth"
	"github.com/bluelock-go/shared/di"
	"github.com/bluelock-go/shared/storage/state/statemanager"
	"github.com/bluelock-go/shared/storage/state/token"
	"github.com/stretchr/testify/assert"
)

func TestHandleRequestWithRetriesWith200StatusCode(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, err := statemanager.NewStateManager(filePath)
	if err != nil {
		t.Errorf("Failed to create StateManager: %v", err)
	}

	tokenStates := map[string]token.TokenState{
		"test-token1": {Status: token.TokenActive},
		"test-token2": {Status: token.TokenActive},
		"test-token3": {Status: token.TokenExhausted},
		"test-token4": {Status: token.TokenUnauthorized},
	}
	for tokenID, tokenState := range tokenStates {
		err = sm.ReplaceTokenState(tokenID, tokenState)
		assert.NoError(t, err)
	}

	// Create a new client
	credentials := []auth.Credential{
		{CredKey: "test-token1"},
		{CredKey: "test-token2"},
		{CredKey: "test-token3"},
		{CredKey: "test-token4"},
	}
	container := di.NewContainer().
		SetLogger(&shared.CustomLogger{Logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}).
		SetStateManager(sm).
		SetCredentials(credentials)
	client := NewClient(container)

	// Define a request callback function
	requestCallback := func(cred *auth.Credential) (*http.Response, error) {
		// Simulate a request and response with a successful status code
		return &http.Response{StatusCode: 200}, nil
	}
	// Call the method under test
	response, err := client.HandleRequestWithRetries(requestCallback)
	// Assert the results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got: %d", response.StatusCode)
	}
}
func TestHandleRequestWithRetriesWithOther2xxStatusCode(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, err := statemanager.NewStateManager(filePath)
	if err != nil {
		t.Errorf("Failed to create StateManager: %v", err)
	}

	tokenStates := map[string]token.TokenState{
		"test-token1": {Status: token.TokenActive},
		"test-token2": {Status: token.TokenActive},
		"test-token3": {Status: token.TokenExhausted},
		"test-token4": {Status: token.TokenUnauthorized},
	}
	for tokenID, tokenState := range tokenStates {
		err = sm.ReplaceTokenState(tokenID, tokenState)
		assert.NoError(t, err)
	}

	// Create a new client
	credentials := []auth.Credential{
		{CredKey: "test-token1"},
		{CredKey: "test-token2"},
		{CredKey: "test-token3"},
		{CredKey: "test-token4"},
	}
	container := di.NewContainer().
		SetLogger(&shared.CustomLogger{Logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}).
		SetStateManager(sm).
		SetCredentials(credentials)
	client := NewClient(container)

	// Define a request callback function
	requestCallback := func(cred *auth.Credential) (*http.Response, error) {
		// Simulate a request and response with a successful status code
		return &http.Response{StatusCode: 201, Body: io.NopCloser(bytes.NewReader([]byte("Created")))}, nil
	}
	// Call the method under test
	response, err := client.HandleRequestWithRetries(requestCallback)
	// Assert the results
	if response != nil {
		t.Errorf("Expected no response, got: %v", response)
	}
	if err == nil {
		t.Errorf("Expected an error for non-200 status code, got nil")
	} else {
		assert.Contains(t, err.Error(), fmt.Sprintf("unexpected 2xx response code: %d for token", 201))
	}
}

func TestHandleRequestWithRetriesWith401StatusCode(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, err := statemanager.NewStateManager(filePath)
	if err != nil {
		t.Errorf("Failed to create StateManager: %v", err)
	}

	tokenStates := map[string]token.TokenState{
		"test-token1": {Status: token.TokenActive},
		"test-token2": {Status: token.TokenActive},
		"test-token4": {Status: token.TokenUnauthorized},
	}
	for tokenID, tokenState := range tokenStates {
		err = sm.ReplaceTokenState(tokenID, tokenState)
		assert.NoError(t, err)
	}

	// Create a new client
	credentials := []auth.Credential{
		{CredKey: "test-token1"},
		{CredKey: "test-token2"},
		{CredKey: "test-token3"},
		{CredKey: "test-token4"},
	}
	container := di.NewContainer().
		SetLogger(&shared.CustomLogger{Logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}).
		SetStateManager(sm).
		SetCredentials(credentials)
	client := NewClient(container)

	requestCallback := func(cred *auth.Credential) (*http.Response, error) {
		// Simulate a request that returns an error
		return &http.Response{StatusCode: 401}, nil
	}
	// Call the method under test
	response, err := client.HandleRequestWithRetries(requestCallback)
	// Assert the results
	if response != nil {
		t.Errorf("Expected response to be nil due to error, got: %v", response)
	}

	if err == nil {
		t.Errorf("Expected an error due to unauthorized access, got nil")
	} else {
		assert.True(t, errors.Is(err, statemanager.ErrAllTokenIgnored))
	}

	for tokenID, tokenState := range sm.State.TokenStates {
		if !tokenState.IsUnauthorized() {
			t.Errorf("Expected the token's status to be %s. but got %s for tokenID: %s", token.TokenUnauthorized, tokenState.Status, tokenID)
		}
	}
}

func TestHandleRequestWithRetriesWith429StatusCode(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, err := statemanager.NewStateManager(filePath)
	if err != nil {
		t.Errorf("Failed to create StateManager: %v", err)
	}

	tokenStates := map[string]token.TokenState{
		"test-token1": {Status: token.TokenActive},
		"test-token2": {Status: token.TokenActive},
		"test-token3": {Status: token.TokenExhausted},
		"test-token4": {Status: token.TokenUnauthorized},
	}
	for tokenID, tokenState := range tokenStates {
		err = sm.ReplaceTokenState(tokenID, tokenState)
		assert.NoError(t, err)
	}

	// Create a new client
	credentials := []auth.Credential{
		{CredKey: "test-token1"},
		{CredKey: "test-token2"},
		{CredKey: "test-token3"},
		{CredKey: "test-token4"},
	}
	container := di.NewContainer().
		SetLogger(&shared.CustomLogger{Logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}).
		SetStateManager(sm).
		SetCredentials(credentials)
	client := NewClient(container)

	// Define a request callback function
	requestCallback := func(cred *auth.Credential) (*http.Response, error) {
		// Simulate a request that returns an error
		return &http.Response{StatusCode: 429}, nil
	}
	// Call the method under test
	response, err := client.HandleRequestWithRetries(requestCallback)
	// Assert the results
	if response != nil {
		t.Errorf("Expected response to be nil due to error, got: %v", response)
	}

	if err == nil {
		t.Errorf("Expected an error due to rate limiting, got nil")
	} else {
		assert.Contains(t, err.Error(), fmt.Sprintf("exceeded maximum reset limit(%d) without a successful response", MAX_ATTEMPTS))
	}
}

func TestHandleRequestWithRetriesWithOther4xxStatusCode(t *testing.T) {
	filePath := "test_state.json"
	defer os.Remove(filePath)

	sm, err := statemanager.NewStateManager(filePath)
	if err != nil {
		t.Errorf("Failed to create StateManager: %v", err)
	}

	tokenStates := map[string]token.TokenState{
		"test-token1": {Status: token.TokenActive},
	}
	for tokenID, tokenState := range tokenStates {
		err = sm.ReplaceTokenState(tokenID, tokenState)
		assert.NoError(t, err)
	}

	// Create a new client
	credentials := []auth.Credential{
		{CredKey: "test-token1"},
	}
	container := di.NewContainer().
		SetLogger(&shared.CustomLogger{Logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}).
		SetStateManager(sm).
		SetCredentials(credentials)
	client := NewClient(container)

	// Define a request callback function
	requestCallback := func(cred *auth.Credential) (*http.Response, error) {
		// Simulate a request that returns an error
		return &http.Response{StatusCode: 404}, nil
	}
	// Call the method under test
	response, err := client.HandleRequestWithRetries(requestCallback)
	// Assert the results
	if response != nil {
		t.Errorf("Expected response to be nil due to error, got: %v", response)
	}

	if err == nil {
		t.Errorf("Expected an error due to rate limiting, got nil")
	} else {
		assert.Contains(t, err.Error(), fmt.Sprintf("unhandled response code: %d for token:", 404))
	}
}
