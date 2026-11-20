package metrics

import (
	"context"
	"runtime"
	"sort"

	"github.com/shirou/gopsutil/v4/process"
)

// processes returns every process, busiest first. A process keeps its object across samples so Percent(0) can
// measure since the last sample; on the first sample every process reads 0.
func (c *Collector) processes(ctx context.Context) []Process {
	pids, err := process.PidsWithContext(ctx)
	if err != nil {
		return nil
	}
	alive := make(map[int32]*process.Process, len(pids))
	list := make([]Process, 0, len(pids))
	cores := float64(runtime.NumCPU())
	for _, pid := range pids {
		p, ok := c.procs[pid]
		if !ok {
			if p, err = process.NewProcessWithContext(ctx, pid); err != nil {
				continue // exited between listing and opening
			}
		}
		alive[pid] = p
		pct, err := p.PercentWithContext(ctx, 0)
		if err != nil {
			continue // no permission or exited
		}
		name, _ := p.NameWithContext(ctx)
		user, _ := p.UsernameWithContext(ctx)
		status, _ := p.StatusWithContext(ctx)
		var memMB float64
		if mi, err := p.MemoryInfoWithContext(ctx); err == nil {
			memMB = float64(mi.RSS) / (1 << 20)
		}
		ppid, _ := p.PpidWithContext(ctx)
		list = append(list, Process{
			PID: pid, Name: name, User: user, Status: processStatus(status),
			CPU: round2(pct / cores), MemoryMB: round2(memMB), PPID: ppid,
		})
	}
	c.procs = alive

	sort.Slice(list, func(i, j int) bool {
		if list[i].CPU != list[j].CPU {
			return list[i].CPU > list[j].CPU
		}
		return list[i].MemoryMB > list[j].MemoryMB
	})
	return list
}

// processStatus maps gopsutil's states to BoxMesh's ProcessStatus.
func processStatus(s []string) string {
	for _, v := range s {
		switch v {
		case process.Stop:
			return "suspended"
		case process.Zombie:
			return "not-responding"
		}
	}
	return "running"
}
