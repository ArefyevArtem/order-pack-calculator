package calculator

import "fmt"

// PackSizesRequest is the body for PUT /api/v1/pack-sizes.
type PackSizesRequest struct {
	Sizes []int `json:"sizes"`
}

// Validate checks non-empty list (duplicates validated in use case).
func (r *PackSizesRequest) Validate() error {
	if len(r.Sizes) == 0 {
		return fmt.Errorf("sizes must not be empty")
	}
	return nil
}

// PackSizesResponse returns current configuration.
type PackSizesResponse struct {
	Sizes []int `json:"sizes"`
}

// CalculateRequest is the body for POST /api/v1/calculate.
type CalculateRequest struct {
	Items int `json:"items"`
}

// Validate checks basic input rules.
func (r *CalculateRequest) Validate() error {
	if r.Items <= 0 {
		return fmt.Errorf("items must be positive")
	}
	return nil
}

// PackLine matches the UI table columns Pack | Quantity with stable JSON ordering.
type PackLine struct {
	Pack     int `json:"pack"`
	Quantity int `json:"quantity"`
}

// CalculateResponse is the JSON envelope for calculate.
type CalculateResponse struct {
	Packs []PackLine `json:"packs"`
}

// ErrorResponse is the JSON body for all API error responses (4xx/5xx).
type ErrorResponse struct {
	Error string `json:"error"`
}
