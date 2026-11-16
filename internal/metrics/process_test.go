package metrics

import (
	"testing"

	"github.com/shirou/gopsutil/v4/process"
)

func TestProcessStatus(t *testing.T) {
	cases := map[string][]string{
		"running":        {process.Running},
		"suspended":      {process.Stop},
		"not-responding": {process.Zombie},
	}
	for want, in := range cases {
		if got := processStatus(in); got != want {
			t.Errorf("processStatus(%v) = %q, want %q", in, got, want)
		}
	}
	if got := processStatus(nil); got != "running" {
		t.Errorf("processStatus(nil) = %q, want running", got)
	}
}
