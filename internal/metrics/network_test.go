package metrics

import "testing"

func TestMbps(t *testing.T) {
	cases := []struct {
		name       string
		prev, cur  uint64
		secs, want float64
	}{
		{"1 MB in 1s is 8 Mbps", 0, 1_000_000, 1, 8},
		{"half over 2s", 1_000_000, 2_000_000, 2, 4},
		{"counter reset", 5_000_000, 100, 10, 0},
		{"no time passed", 0, 100, 0, 0},
	}
	for _, c := range cases {
		if got := mbps(c.prev, c.cur, c.secs); got != c.want {
			t.Errorf("%s: mbps(%d, %d, %v) = %v, want %v", c.name, c.prev, c.cur, c.secs, got, c.want)
		}
	}
}
