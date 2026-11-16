package metrics

import (
	"context"
	"net"
	"time"

	gnet "github.com/shirou/gopsutil/v4/net"
)

// netCounter remembers the last byte counters so the next sample can turn them into a rate.
type netCounter struct {
	at         time.Time
	recv, sent uint64
}

// rates returns Mbps in and out since the previous call; 0, 0 on the first call or if counters can't be read.
func (n *netCounter) rates(ctx context.Context, now time.Time) (in, out float64) {
	io, err := gnet.IOCountersWithContext(ctx, false)
	if err != nil || len(io) == 0 {
		return 0, 0
	}
	if !n.at.IsZero() {
		secs := now.Sub(n.at).Seconds()
		in, out = mbps(n.recv, io[0].BytesRecv, secs), mbps(n.sent, io[0].BytesSent, secs)
	}
	n.at, n.recv, n.sent = now, io[0].BytesRecv, io[0].BytesSent
	return in, out
}

// mbps is the rate between two byte counters; a counter that went backwards (reset or wrapped) reads as 0.
func mbps(prev, cur uint64, secs float64) float64 {
	if cur < prev || secs <= 0 {
		return 0
	}
	return float64(cur-prev) * 8 / secs / 1e6
}

// interfaces lists the network cards that are up, loopback excluded.
func interfaces() []Interface {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []Interface
	for _, i := range ifs {
		if i.Flags&net.FlagLoopback != 0 || i.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, _ := i.Addrs()
		var ips []string
		for _, a := range addrs {
			if n, ok := a.(*net.IPNet); ok {
				ips = append(ips, n.IP.String())
			}
		}
		out = append(out, Interface{Name: i.Name, MAC: i.HardwareAddr.String(), IPs: ips})
	}
	return out
}
