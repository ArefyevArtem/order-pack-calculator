package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMinPacks_OK: valid inputs where an exact packing exists; asserts optimal map, sum invariant, optional total pack count.
func TestMinPacks_OK(t *testing.T) {
	cases := []struct {
		name          string
		sizes         []int
		target        int
		want          map[int]int
		long          bool
		wantPackCount int
	}{
		{
			// Fewest packs: one 100-pack beats five 20-packs (etc.).
			name:   "one_large_pack_beats_many_small",
			sizes:  []int{25, 50, 100},
			target: 100,
			want:   map[int]int{100: 1},
		},
		{
			// Optimum is not “largest first”: 60 → three 20s, not 50+10.
			name:   "min_total_packs_not_greedy_by_face_value",
			sizes:  []int{20, 50, 100},
			target: 60,
			want:   map[int]int{20: 3},
		},
		{
			// Single denomination; unlimited copies.
			name:   "single_size_repeated",
			sizes:  []int{7},
			target: 21,
			want:   map[int]int{7: 3},
		},
		{
			// Classic “fewest coins” optimum (here two 3s, not six 1s).
			name:   "coin_change_fewest_coins",
			sizes:  []int{1, 3, 4},
			target: 6,
			want:   map[int]int{3: 2},
		},
		{
			// Mix of small/medium sizes to hit target with minimum pack count.
			name:   "mixed_denominations",
			sizes:  []int{1, 5, 10},
			target: 27,
			want:   map[int]int{10: 2, 5: 1, 1: 2},
		},
		{
			// Target equals one allowed pack size exactly once.
			name:   "exact_single_denomination",
			sizes:  []int{42},
			target: 42,
			want:   map[int]int{42: 1},
		},
		{
			// Duplicate values in sizes slice must be deduplicated internally.
			name:   "duplicate_sizes_in_input_ignored",
			sizes:  []int{10, 10, 5, 5},
			target: 25,
			want:   map[int]int{10: 2, 5: 1},
		},
		{
			// Another case where greedy-by-value fails; 5+3+3 beats alternatives.
			name:   "another_non_greedy",
			sizes:  []int{2, 3, 5},
			target: 11,
			want:   map[int]int{5: 1, 3: 2},
		},
		{
			// Powers of two: one of each pack sums to 15 with four packs total.
			name:   "powers_of_two_style",
			sizes:  []int{1, 2, 4, 8},
			target: 15,
			want:   map[int]int{8: 1, 4: 1, 2: 1, 1: 1},
		},
		{
			// Large target: O(target·n) DP stress; known split + min total pack count (skip under -short).
			name:          "recruiter_edge_500k",
			sizes:         []int{23, 31, 53},
			target:        500_000,
			want:          map[int]int{23: 2, 31: 7, 53: 9429},
			long:          true,
			wantPackCount: 9438, // 2+7+9429
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.long && testing.Short() {
				t.Skip("heavy DP; run without -short")
			}
			got, err := MinPacks(tt.sizes, tt.target)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assertPackingSumsToTarget(t, tt.target, got)
			if tt.wantPackCount != 0 {
				assert.Equal(t, tt.wantPackCount, totalPackCount(got), "total pack count")
			}
		})
	}
}

// TestMinPacks_ErrNoExactPacking: target cannot be expressed as a non-negative integer combination of sizes.
func TestMinPacks_ErrNoExactPacking(t *testing.T) {
	cases := []struct {
		name   string
		sizes  []int
		target int
	}{
		// 5 is not a combination of 4 and 6.
		{name: "no_linear_combination", sizes: []int{4, 6}, target: 5},
		// Multiples of 5 only; 3 unreachable.
		{name: "coprime_gap", sizes: []int{5}, target: 3},
		// Even pack sizes cannot sum to an odd target.
		{name: "even_only_odd_target", sizes: []int{2, 4}, target: 7},
		// Target below smallest pack.
		{name: "too_small_set", sizes: []int{10, 20}, target: 5},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MinPacks(tt.sizes, tt.target)
			assert.ErrorIs(t, err, ErrNoExactPacking)
		})
	}
}

// TestMinPacks_ContractErrors: invalid arguments; expect plain errors, not ErrNoExactPacking.
func TestMinPacks_ContractErrors(t *testing.T) {
	cases := []struct {
		name       string
		sizes      []int
		target     int
		wantSubstr string
	}{
		{name: "target_zero", sizes: []int{1}, target: 0, wantSubstr: "target must be positive"},
		{name: "target_negative", sizes: []int{1}, target: -1, wantSubstr: "target must be positive"},
		// No denominations provided.
		{name: "empty_sizes", sizes: []int{}, target: 10, wantSubstr: "no pack sizes"},
		// After filtering non-positive sizes, nothing usable remains.
		{name: "no_positive_sizes", sizes: []int{0, -3}, target: 10, wantSubstr: "no positive pack sizes"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MinPacks(tt.sizes, tt.target)
			require.Error(t, err)
			assert.False(t, errors.Is(err, ErrNoExactPacking))
			assert.ErrorContains(t, err, tt.wantSubstr)
		})
	}
}

// assertPackingSumsToTarget checks sum(size*qty) == target (redundant with exact want map but catches reconstruction bugs).
func assertPackingSumsToTarget(t *testing.T, target int, got map[int]int) {
	t.Helper()
	sum := 0
	for s, n := range got {
		sum += s * n
	}
	assert.Equal(t, target, sum, "sum(size*qty) must equal target")
}

// totalPackCount returns sum of quantities (number of physical packs in the breakdown).
func totalPackCount(m map[int]int) int {
	n := 0
	for _, c := range m {
		n += c
	}
	return n
}
