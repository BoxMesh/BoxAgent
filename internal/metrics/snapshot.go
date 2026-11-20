// Package metrics samples a machine's identity, usage and processes.
package metrics

import "time"

// Snapshot mirrors Device, MetricPoint and DeviceProcess in BoxMesh/src/types/device.ts.
type Snapshot struct {
	At             time.Time   `json:"at"`
	Host           string      `json:"host"`
	OS             string      `json:"os"`
	OSVersion      string      `json:"osVersion"`
	OSArchitecture string      `json:"osArchitecture"`
	OSFamily       string      `json:"osFamily"`
	AgentVersion   string      `json:"agentVersion"`
	Interfaces     []Interface `json:"interfaces"`
	Hardware       Hardware    `json:"hardware"`
	// Usage in percent; network in Mbps since the previous sample.
	CPU       float64   `json:"cpu"`
	RAM       float64   `json:"ram"`
	Disk      float64   `json:"disk"`
	NetIn     float64   `json:"netIn"`
	NetOut    float64   `json:"netOut"`
	Processes []Process `json:"processes"`
}

type Interface struct {
	Name string   `json:"name"`
	MAC  string   `json:"mac"`
	IPs  []string `json:"ips"`
}

type Hardware struct {
	CPU           string `json:"cpu"`
	Cores         int    `json:"cores"`
	RAMBytes      uint64 `json:"ramBytes"`
	DiskBytes     uint64 `json:"diskBytes"`
	UptimeSeconds uint64 `json:"uptimeSeconds"`
}

type Process struct {
	PID      int32   `json:"pid"`
	Name     string  `json:"name"`
	User     string  `json:"user"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu"` // percent of the whole machine
	MemoryMB float64 `json:"memoryMb"`
	PPID     int32   `json:"ppid"` // parent process; 0 when unknown
}
