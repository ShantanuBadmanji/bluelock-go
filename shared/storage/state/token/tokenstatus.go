package token

import "slices"

// TokenStatus defines possible token statuses
type TokenStatus string

const (
	TokenActive       TokenStatus = "active"
	TokenExhausted    TokenStatus = "exhausted"
	TokenUnauthorized TokenStatus = "unauthorized"
)

var ValidTokenStatuses = []TokenStatus{TokenActive, TokenExhausted, TokenUnauthorized}

func IsTokenStatusValid(status TokenStatus) bool {
	return slices.Contains(ValidTokenStatuses, status)
}
