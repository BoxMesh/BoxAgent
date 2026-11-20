// Package report sends snapshots to BoxServer over gRPC (the Ping RPC in proto/device.proto).
package report

import (
	"context"
	"net"
	"slices"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"boxagent/internal/metrics"
	"boxagent/internal/pb"
)

type Client struct {
	addr string
	id   string
	conn *grpc.ClientConn
	rpc  pb.DeviceGRPCClient
}

// Dial prepares a client for addr (host:port). gRPC connects on first use and reconnects on its own.
func Dial(addr, id string) (*Client, error) {
	// shortcut: plaintext, as BoxServer's gRPC port has no TLS yet; switch to TLS credentials when it does.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{addr: addr, id: id, conn: conn, rpc: pb.NewDeviceGRPCClient(conn)}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

// Send reports one snapshot, giving up after 5 seconds so a down server never stalls sampling.
func (c *Client) Send(ctx context.Context, s *metrics.Snapshot) error {
	ip, mac := route(c.addr, s.Interfaces)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := c.rpc.Ping(ctx, request(c.id, s, ip, mac))
	return err
}

func request(id string, s *metrics.Snapshot, ip, mac string) *pb.DeviceRequest {
	procs := make([]*pb.DeviceProcess, len(s.Processes))
	for i, p := range s.Processes {
		procs[i] = &pb.DeviceProcess{Pid: p.PID, Name: p.Name, User: p.User, Status: p.Status, Cpu: p.CPU, MemoryMb: p.MemoryMB, Ppid: p.PPID}
	}
	return &pb.DeviceRequest{
		Processes:      procs,
		Id:             id,
		HostName:       s.Host,
		OsName:         s.OS,
		OsVersion:      s.OSVersion,
		OsArchitecture: s.OSArchitecture,
		IpAddress:      ip,
		MacAddress:     mac,
		Cpu:            s.Hardware.CPU,
		Ram:            int64(s.Hardware.RAMBytes),
		Disk:           int64(s.Hardware.DiskBytes),
		Uptime:         int64(s.Hardware.UptimeSeconds),
		AgentVersion:   s.AgentVersion,
		CpuUsage:       s.CPU,
		RamUsage:       s.RAM,
		DiskUsage:      s.Disk,
		NetIn:          s.NetIn,
		NetOut:         s.NetOut,
	}
}

// route returns the IP and MAC of the card the OS would use to reach addr, so the server records the address
// it can reach this machine on (on a Vagrant VM that is the host-only card, not the NAT one).
func route(addr string, ifs []metrics.Interface) (ip, mac string) {
	conn, err := net.Dial("udp", addr) // a UDP "dial" only consults the routing table; nothing is sent
	if err != nil {
		return "", ""
	}
	defer conn.Close()
	ip = conn.LocalAddr().(*net.UDPAddr).IP.String()
	for _, i := range ifs {
		if slices.Contains(i.IPs, ip) {
			return ip, i.MAC
		}
	}
	return ip, ""
}
