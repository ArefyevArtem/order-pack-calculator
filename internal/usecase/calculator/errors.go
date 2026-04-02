package calculator

import "errors"

// Use-case / input errors (not pure DP). HTTP layer can map them to 400.
var (
	ErrNoPackSizes      = errors.New("no pack sizes configured")
	ErrItemsNotPositive = errors.New("items amount must be positive")
)
