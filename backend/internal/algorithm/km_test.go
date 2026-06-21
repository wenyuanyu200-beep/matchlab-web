package algorithm

import "testing"

func TestKMOutperformsRowGreedy(t *testing.T) {
	weights := [][]int{{9, 8, 1}, {8, 1, 1}, {1, 8, 8}}
	pairs, kmTotal, err := MaximumWeightMatching(weights)
	if err != nil || kmTotal != 24 {
		t.Fatalf("KM: pairs=%v total=%d err=%v", pairs, kmTotal, err)
	}
	if len(pairs) != 3 {
		t.Fatalf("KM returned %d pairs, want 3", len(pairs))
	}
	_, greedyTotal, err := GreedyMatching(weights)
	if err != nil || greedyTotal != 18 {
		t.Fatalf("greedy: total=%d err=%v", greedyTotal, err)
	}
}

func TestMaximumWeightMatchingSupportsRectangularMatrices(t *testing.T) {
	tests := []struct {
		name    string
		weights [][]int
		total   int
		pairs   int
	}{
		{name: "more columns", weights: [][]int{{5, 1, 4}, {4, 6, 2}}, total: 11, pairs: 2},
		{name: "more rows", weights: [][]int{{9, 1}, {8, 7}, {6, 5}}, total: 16, pairs: 2},
		{name: "negative", weights: [][]int{{-1, -2}, {-3, -4}}, total: -5, pairs: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pairs, total, err := MaximumWeightMatching(test.weights)
			if err != nil || total != test.total || len(pairs) != test.pairs {
				t.Fatalf("pairs=%v total=%d err=%v, want pairs=%d total=%d", pairs, total, err, test.pairs, test.total)
			}
		})
	}
}

func TestMatchingHandlesEmptyAndRejectsRaggedMatrices(t *testing.T) {
	for _, function := range []func([][]int) ([]Pair, int, error){MaximumWeightMatching, GreedyMatching} {
		pairs, total, err := function(nil)
		if err != nil || len(pairs) != 0 || total != 0 {
			t.Fatalf("empty: pairs=%v total=%d err=%v", pairs, total, err)
		}
		if _, _, err := function([][]int{{1, 2}, {3}}); err == nil {
			t.Fatal("ragged matrix should return an error")
		}
	}
}
