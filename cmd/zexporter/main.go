// ZExporter samples this machine's metrics and reports each sample to BoxServer over gRPC.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"boxagent/internal/identity"
	"boxagent/internal/metrics"
	"boxagent/internal/report"

	"github.com/spf13/cobra"
)

// Set at build time: -ldflags "-X main.version=..."
var version = "0.0.0"

func main() {
	var server string
	var interval time.Duration
	cmd := &cobra.Command{
		Use:           "zexporter",
		Short:         "Report this machine's metrics to BoxServer",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), server, interval)
		},
	}
	cmd.Flags().StringVarP(&server, "server", "s", "", "BoxServer gRPC address, host:port, e.g. 192.168.1.10:8081")
	cmd.Flags().DurationVarP(&interval, "interval", "i", 10*time.Second, "time between samples")
	cmd.MarkFlagRequired("server")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "zexporter:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, server string, interval time.Duration) error {
	id, err := identity.LoadOrCreate(idFile())
	if err != nil {
		return fmt.Errorf("device id: %w", err)
	}
	reporter, err := report.Dial(server, id)
	if err != nil {
		return fmt.Errorf("server %s: %w", server, err)
	}
	defer reporter.Close()
	log.Printf("reporting to %s as device %s", server, id)

	c := metrics.NewCollector(version)
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		if s, err := c.Collect(ctx); err != nil {
			log.Printf("collect: %v", err)
		} else if err := reporter.Send(ctx, s); err != nil {
			// A failed report is logged and dropped; the next sample tries again.
			log.Printf("report: %v", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
	}
}

func idFile() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "ZExporter", "id")
	}
	return "/var/lib/zexporter/id"
}
