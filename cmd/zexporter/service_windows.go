//go:build windows

package main

import (
	"context"

	"golang.org/x/sys/windows/svc"
)

type service struct{ run func(ctx context.Context) error }

// Execute is what the Service Control Manager drives: report Running, run the agent, and cancel it on Stop/Shutdown.
func (s service) Execute(_ []string, req <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	status <- svc.Status{State: svc.StartPending}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.run(ctx) }()
	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case c := <-req:
			if c.Cmd == svc.Stop || c.Cmd == svc.Shutdown {
				status <- svc.Status{State: svc.StopPending}
				cancel()
				<-done
				return false, 0
			}
		case <-done:
			return true, 1 // service-specific error code, so the failure actions restart it
		}
	}
}

// runService reports whether the process was started by the Service Control Manager, and if so runs it to the end.
func runService(run func(ctx context.Context) error) (bool, error) {
	if is, err := svc.IsWindowsService(); err != nil || !is {
		return false, err
	}
	return true, svc.Run("ZExporter", service{run})
}
