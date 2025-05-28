package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/victorximenis/multitenant/core"
)

type Worker struct {
	resolver     *TenantResolver
	processAll   bool
	tenantName   string
	pollInterval time.Duration
	shutdownChan chan struct{}
	shutdownDone chan struct{}
}

type WorkerConfig struct {
	TenantService core.TenantService
	ProcessAll    bool
	TenantName    string
	EnvVarName    string
	PollInterval  time.Duration
}

func NewWorker(config WorkerConfig) *Worker {
	if config.PollInterval == 0 {
		config.PollInterval = 1 * time.Minute
	}

	return &Worker{
		resolver:     NewTenantResolver(config.TenantService, config.EnvVarName),
		processAll:   config.ProcessAll,
		tenantName:   config.TenantName,
		pollInterval: config.PollInterval,
		shutdownChan: make(chan struct{}),
		shutdownDone: make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context, processFn func(context.Context) error) error {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		w.Shutdown()
	}()

	// Start worker loop
	go w.run(ctx, processFn)

	return nil
}

func (w *Worker) run(ctx context.Context, processFn func(context.Context) error) {
	defer close(w.shutdownDone)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Process immediately on start
	w.process(ctx, processFn)

	for {
		select {
		case <-ticker.C:
			w.process(ctx, processFn)
		case <-w.shutdownChan:
			return
		}
	}
}

func (w *Worker) process(ctx context.Context, processFn func(context.Context) error) {
	if w.processAll {
		// Process all tenants
		err := w.resolver.ForEachTenant(ctx, processFn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing tenants: %v\n", err)
		}
	} else {
		// Process single tenant
		tenantName := w.tenantName
		if tenantName == "" {
			// Try to get from environment
			tenantName = os.Getenv(w.resolver.envVarName)
		}

		if tenantName == "" {
			fmt.Fprintf(os.Stderr, "No tenant specified\n")
			return
		}

		err := w.resolver.WithTenant(ctx, tenantName, processFn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing tenant %s: %v\n", tenantName, err)
		}
	}
}

func (w *Worker) Shutdown() {
	close(w.shutdownChan)
	<-w.shutdownDone
}
