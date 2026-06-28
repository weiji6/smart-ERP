package erp

import "testing"

func TestAdSelectionChances(t *testing.T) {
	cases := map[Money]int{0: 0, 1: 1, 2: 1, 3: 2, 5: 3}
	for ad, want := range cases {
		if got := AdSelectionChances(ad); got != want {
			t.Fatalf("%dW 广告选单机会 want %d got %d", ad, want, got)
		}
	}
}

func TestPredictAdRanking(t *testing.T) {
	got := PredictAdRanking(3, []Money{5, 3, 2, 1})
	if got != 2 {
		t.Fatalf("只有 5W 高于 3W，排名应为 2，got %d", got)
	}
}
