package report

import (
	"testing"

	"boxagent/internal/metrics"
)

func TestRoutePicksTheCardThatReachesTheServer(t *testing.T) {
	ifs := []metrics.Interface{
		{Name: "eth0", MAC: "08:00:27:00:00:01", IPs: []string{"10.0.2.15"}},
		{Name: "lo", MAC: "00:00:00:00:00:00", IPs: []string{"127.0.0.1"}},
	}
	ip, mac := route("127.0.0.1:8081", ifs)
	if ip != "127.0.0.1" || mac != "00:00:00:00:00:00" {
		t.Fatalf("route = %q, %q; want 127.0.0.1 and lo's MAC", ip, mac)
	}
}

func TestRequestCarriesTheSnapshot(t *testing.T) {
	s := &metrics.Snapshot{
		Host: "debian-1", OS: "debian", OSVersion: "13.1", OSArchitecture: "x86_64", AgentVersion: "v0.1.3",
		Hardware: metrics.Hardware{CPU: "Test CPU", RAMBytes: 512, DiskBytes: 1024, UptimeSeconds: 60},
		CPU:      12.5, RAM: 40, Disk: 3.2, NetIn: 1.5, NetOut: 0.25,
	}
	r := request("id-1", s, "192.168.56.11", "08:00:27:28:cb:85")
	if r.Id != "id-1" || r.HostName != "debian-1" || r.IpAddress != "192.168.56.11" || r.MacAddress != "08:00:27:28:cb:85" ||
		r.Ram != 512 || r.Disk != 1024 || r.Uptime != 60 || r.AgentVersion != "v0.1.3" ||
		r.CpuUsage != 12.5 || r.RamUsage != 40 || r.DiskUsage != 3.2 || r.NetIn != 1.5 || r.NetOut != 0.25 {
		t.Fatalf("request lost a field: %v", r)
	}
}
