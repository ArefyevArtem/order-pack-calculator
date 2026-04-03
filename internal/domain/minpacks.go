package domain

import (
	"errors"
	"math"
	"sort"
)

// ErrNoExactPacking is returned when no combination of the given sizes sums to the target exactly.
var ErrNoExactPacking = errors.New("cannot reach exact amount with given pack sizes")

// MinPacks: minimum count of unlimited pack sizes that sum to target exactly.
// Preconditions: target > 0, non-empty positive sizes — else generic error, not ErrNoExactPacking.
// Bottom-up DP over 1..target, then reconstruct via parent pointers. O(target * len(sizes)) time, O(target) space.
func MinPacks(sizes []int, target int) (map[int]int, error) {
	if target <= 0 {
		return nil, errors.New("target must be positive")
	}
	if len(sizes) == 0 {
		return nil, errors.New("no pack sizes")
	}

	uniq := make([]int, 0, len(sizes))
	seen := make(map[int]struct{})
	for _, s := range sizes {
		if s <= 0 {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		uniq = append(uniq, s)
	}
	if len(uniq) == 0 {
		return nil, errors.New("no positive pack sizes")
	}
	sort.Ints(uniq)

	dp := make([]int, target+1)
	parent := make([]int, target+1)
	for i := 1; i <= target; i++ {
		dp[i] = math.MaxInt32
		parent[i] = -1
	}

	for a := 1; a <= target; a++ {
		for _, s := range uniq {
			if a < s {
				continue
			}
			if dp[a-s] == math.MaxInt32 {
				continue
			}
			if cand := dp[a-s] + 1; cand < dp[a] {
				dp[a] = cand
				parent[a] = s
			}
		}
	}

	if dp[target] == math.MaxInt32 {
		return nil, ErrNoExactPacking
	}

	counts := make(map[int]int)
	for cur := target; cur > 0; {
		s := parent[cur]
		if s <= 0 {
			return nil, ErrNoExactPacking
		}
		counts[s]++
		cur -= s
	}
	return counts, nil
}
