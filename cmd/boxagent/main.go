// BoxAgent samples this machine's metrics and prints one JSON line per sample to stdout.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"boxagent/internal/metrics"
)

// Set at build time: -ldflags "-X main.version=..."
var version = "dev"

func main() {
	interval := flag.Duration("interval", 10*time.Second, "time between samples")
	top := flag.Int("top", 10, "number of processes to report, busiest first")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := metrics.NewCollector(*top, version)
	out := json.NewEncoder(os.Stdout)
	tick := time.NewTicker(*interval)
	defer tick.Stop()
	for {
		if s, err := c.Collect(ctx); err != nil {
			log.Printf("collect: %v", err)
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
