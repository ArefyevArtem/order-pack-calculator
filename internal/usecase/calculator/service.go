package calculator

import (
	"context"
	"fmt"
	"sort"

	"order-pack-calculator/internal/domain"
)

// PackStore abstracts persistence of the configured pack size list (e.g. PostgreSQL or a test double).
type PackStore interface {
	List(ctx context.Context) ([]int, error)
	ReplaceAll(ctx context.Context, sizes []int) error
}

// Service orchestrates pack persistence and optimal counts.
type Service struct {
	packs PackStore
}

// NewService wires persistence.
func NewService(packs PackStore) *Service {
	return &Service{packs: packs}
}

// Calculate loads the current pack sizes from the DB, then runs MinPacks (fewest packs total).
func (s *Service) Calculate(ctx context.Context, items int) (*Result, error) {
	if items <= 0 {
		return nil, ErrItemsNotPositive
	}

	sizes, err := s.packs.List(ctx)
	if err != nil {
		return nil, err
	}
	if len(sizes) == 0 {
		return nil, ErrNoPackSizes
	}

	counts, err := domain.MinPacks(sizes, items)
	if err != nil {
		return nil, err
	}

	return &Result{Packs: breakdownSorted(counts)}, nil
}

// ReplacePackSizes replaces the full configured size list (what “Submit pack sizes” does in the UI).
func (s *Service) ReplacePackSizes(ctx context.Context, sizes []int) error {
	if err := validatePackSizes(sizes); err != nil {
		return err
	}
	uniq := dedupeSorted(sizes)
	return s.packs.ReplaceAll(ctx, uniq)
}

// ListPackSizes returns current sizes from persistence, ascending.
func (s *Service) ListPackSizes(ctx context.Context) ([]int, error) {
	return s.packs.List(ctx)
}

func validatePackSizes(sizes []int) error {
	if len(sizes) == 0 {
		return ErrNoPackSizes
	}
	seen := make(map[int]struct{}, len(sizes))
	for _, s := range sizes {
		if s <= 0 {
			return fmt.Errorf("pack size must be positive, got %d", s)
		}
		if _, ok := seen[s]; ok {
			return fmt.Errorf("duplicate pack size %d", s)
		}
		seen[s] = struct{}{}
	}
	return nil
}

func dedupeSorted(sizes []int) []int {
	seen := make(map[int]struct{}, len(sizes))
	out := make([]int, 0, len(sizes))
	for _, s := range sizes {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Ints(out)
	return out
}
