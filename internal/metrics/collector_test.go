package metrics

import (
	"context"
	"testing"
	"time"
)

// Runs against the real machine the tests run on.
func TestCollectOnThisMachine(t *testing.T) {
	c := NewCollector(5, "test")
	ctx := context.Background()
	if _, err := c.Collect(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	s, err := c.Collect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if s.Host == "" || s.OSFamily == "" || s.Hardware.RAMBytes == 0 || s.Hardware.DiskBytes == 0 {
		t.Errorf("missing machine info: %+v", s)
	}
	if s.AgentVersion != "test" {
		t.Errorf("agentVersion = %q, want test", s.AgentVersion)
	}
	for name, v := range map[string]float64{"cpu": s.CPU, "ram": s.RAM, "disk": s.Disk} {
		if v < 0 || v > 100 {
			t.Errorf("%s = %v, want 0..100", name, v)
		}
	}
	if s.NetIn < 0 || s.NetOut < 0 {
		t.Errorf("net = %v / %v, want >= 0", s.NetIn, s.NetOut)
	}
	if len(s.Processes) == 0 || len(s.Processes) > 5 {
		t.Errorf("got %d processes, want 1..5", len(s.Processes))
	}
	for _, p := range s.Processes {
		if p.CPU < 0 || p.CPU > 100 {
			t.Errorf("process %d cpu = %v, want 0..100", p.PID, p.CPU)
		}
	}
}
