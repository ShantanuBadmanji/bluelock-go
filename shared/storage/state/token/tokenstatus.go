package token

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/bluelock-go/shared/customerrors"
)

// TokenStatus defines possible token statuses
type TokenStatus string

const (
	TokenActive       TokenStatus = "active"
	TokenExhausted    TokenStatus = "exhausted"
	TokenUnauthorized TokenStatus = "unauthorized"
)

var ValidTokenStatuses = []TokenStatus{TokenActive, TokenExhausted, TokenUnauthorized}
var IgnoredTokenStatuses = []TokenStatus{TokenUnauthorized}

func IsTokenStatusValid(status TokenStatus) bool {
	return slices.Contains(ValidTokenStatuses, status)
}

var (
	ErrUnExpectedTokenStatus = fmt.Errorf("unexpected token status, valid statuses are: %v: %w", ValidTokenStatuses, customerrors.ErrCritical)
)

// MarshalJSON implements the json.Marshaler interface
func (ts TokenStatus) MarshalJSON() ([]byte, error) {
	if !IsTokenStatusValid(ts) {
		return nil, fmt.Errorf("invalid token status: %v. %w", ts, ErrUnExpectedTokenStatus)
	}
	return json.Marshal(string(ts))
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (ts *TokenStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	status := TokenStatus(s)
	if !IsTokenStatusValid(status) {
		return fmt.Errorf("invalid token status: %v. %w", s, ErrUnExpectedTokenStatus)
	}

	*ts = status
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface
func (ts TokenStatus) MarshalText() ([]byte, error) {
	if !IsTokenStatusValid(ts) {
		return nil, fmt.Errorf("invalid token status: %v. %w", ts, ErrUnExpectedTokenStatus)
	}
	return []byte(ts), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (ts *TokenStatus) UnmarshalText(text []byte) error {
	status := TokenStatus(text)
	if !IsTokenStatusValid(status) {
		return fmt.Errorf("invalid token status: %s. %w", text, ErrUnExpectedTokenStatus)
	}

	*ts = status
	return nil
}
