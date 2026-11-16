// BoxAgent samples this machine's metrics and reports each sample to BoxServer over gRPC. Without -server it
// prints one JSON line per sample to stdout instead.
package main

import (
	"context"
	"encoding/json"
	"flag"
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
)

// Set at build time: -ldflags "-X main.version=..."
var version = "dev"

func main() {
	interval := flag.Duration("interval", 10*time.Second, "time between samples")
	top := flag.Int("top", 10, "number of processes to report, busiest first")
	server := flag.String("server", "", "BoxServer gRPC address, host:port; empty keeps samples local")
	idFile := flag.String("id-file", defaultIDFile(), "file that keeps this machine's device ID")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var reporter *report.Client
	if *server != "" {
		id, err := identity.LoadOrCreate(*idFile)
		if err != nil {
			log.Fatalf("device id: %v", err)
		}
		if reporter, err = report.Dial(*server, id); err != nil {
			log.Fatalf("server %s: %v", *server, err)
		}
		defer reporter.Close()
		log.Printf("reporting to %s as device %s", *server, id)
	}

	c := metrics.NewCollector(*top, version)
	out := json.NewEncoder(os.Stdout)
	tick := time.NewTicker(*interval)
	defer tick.Stop()
	for {
		if s, err := c.Collect(ctx); err != nil {
			log.Printf("collect: %v", err)
		} else if reporter != nil {
			// A failed report is logged and dropped; the next sample tries again.
			if err := reporter.Send(ctx, s); err != nil {
				log.Printf("report: %v", err)
			}
		} else if err := out.Encode(s); err != nil {
			log.Printf("write: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

func defaultIDFile() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "BoxAgent", "id")
	}
	return "/var/lib/boxagent/id"
}
