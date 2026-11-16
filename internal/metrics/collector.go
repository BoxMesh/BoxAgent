package metrics

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

// Collector keeps what rates need between samples: network counters and per-process CPU times.
// Not safe for concurrent use.
type Collector struct {
	top          int
	agentVersion string
	diskMount    string
	cpuModel     string
	net          netCounter
	procs        map[int32]*process.Process
}

// NewCollector reports the top busiest processes and tags every snapshot with agentVersion.
func NewCollector(top int, agentVersion string) *Collector {
	mount := "/"
	if runtime.GOOS == "windows" {
		mount = `C:\`
	}
	return &Collector{top: top, agentVersion: agentVersion, diskMount: mount, procs: map[int32]*process.Process{}}
}

// Collect takes one sample. Rates (CPU, network, per-process CPU) cover the time since the previous call;
// on the first call network and process CPU read 0.
func (c *Collector) Collect(ctx context.Context) (*Snapshot, error) {
	now := time.Now()
	h, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("host: %w", err)
	}
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("memory: %w", err)
	}
	du, err := disk.UsageWithContext(ctx, c.diskMount)
	if err != nil {
		return nil, fmt.Errorf("disk: %w", err)
	}
	// Interval 0 measures since the previous call.
	cpuPct, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil || len(cpuPct) == 0 {
		return nil, fmt.Errorf("cpu: %v", err)
	}
	if c.cpuModel == "" {
		if info, err := cpu.InfoWithContext(ctx); err == nil && len(info) > 0 {
			c.cpuModel = info[0].ModelName
		}
	}
	netIn, netOut := c.net.rates(ctx, now)

	return &Snapshot{
		At:             now.UTC(),
		Host:           h.Hostname,
		OS:             h.Platform,
		OSVersion:      h.PlatformVersion,
		OSArchitecture: h.KernelArch,
		OSFamily:       osFamily(runtime.GOOS),
		AgentVersion:   c.agentVersion,
		Interfaces:     interfaces(),
		Hardware: Hardware{
			CPU:           c.cpuModel,
			Cores:         runtime.NumCPU(),
			RAMBytes:      vm.Total,
			DiskBytes:     du.Total,
			UptimeSeconds: h.Uptime,
		},
		CPU:       round2(cpuPct[0]),
		RAM:       round2(vm.UsedPercent),
		Disk:      round2(du.UsedPercent),
		NetIn:     round2(netIn),
		NetOut:    round2(netOut),
		Processes: c.processes(ctx),
	}, nil
}

func osFamily(goos string) string {
	switch goos {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	}
	return "Linux"
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
