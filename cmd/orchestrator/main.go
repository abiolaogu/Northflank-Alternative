// Package main is the entry point for the Platform Orchestrator.
// It initializes all components and starts the HTTP server.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openpaas/platform-orchestrator/internal/adapters/argocd"
	"github.com/openpaas/platform-orchestrator/internal/adapters/coolify"
	"github.com/openpaas/platform-orchestrator/internal/adapters/rancher"
	"github.com/openpaas/platform-orchestrator/internal/api"
	"github.com/openpaas/platform-orchestrator/internal/config"
	"github.com/openpaas/platform-orchestrator/internal/domain"
	"github.com/openpaas/platform-orchestrator/internal/eventbus"
	"github.com/openpaas/platform-orchestrator/internal/repository"
	"github.com/openpaas/platform-orchestrator/internal/workflow"
	"github.com/openpaas/platform-orchestrator/pkg/logger"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	migrate := flag.Bool("migrate", false, "Run database migrations")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Platform Orchestrator\n")
		fmt.Printf("  Version:    %s\n", version)
		fmt.Printf("  Commit:     %s\n", commit)
		fmt.Printf("  Build Date: %s\n", buildDate)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.Logging.Level, cfg.Logging.Format, os.Stdout)
	log.Info().
		Str("version", version).
		Str("environment", cfg.Logging.Level).
		Msg("Starting Platform Orchestrator")

	// Create root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := repository.NewPostgresDB(ctx, &cfg.Database, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations if requested
	if *migrate {
		log.Info().Msg("Running database migrations...")
		if err := db.Migrate(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to run migrations")
		}
		log.Info().Msg("Migrations completed successfully")
		if flag.NArg() == 0 {
			os.Exit(0)
		}
	}

	// Initialize repositories
	projectRepo := repository.NewProjectRepository(db)
	serviceRepo := repository.NewServiceRepository(db)

	// Initialize event bus
	bus, err := eventbus.NewNATSEventBus(&cfg.NATS, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer bus.Close()

	// Initialize adapters
	coolifyAdapter := coolify.NewAdapter(&cfg.Coolify, log)
	rancherAdapter := rancher.NewAdapter(&cfg.Rancher, log)
	argocdAdapter := argocd.NewAdapter(&cfg.ArgoCD, log)

	// Authenticate with ArgoCD if configured
	if cfg.ArgoCD.Username != "" || cfg.ArgoCD.Token != "" {
		if err := argocdAdapter.Authenticate(ctx); err != nil {
			log.Warn().Err(err).Msg("Failed to authenticate with ArgoCD")
		}
	}

	// Initialize workflow engine
	stateMachine := workflow.NewStateMachine(coolifyAdapter, argocdAdapter, bus, serviceRepo, log)

	// Start workflow cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stateMachine.CleanupOldWorkflows(24 * time.Hour)
			}
		}
	}()

	// Subscribe to events for workflow processing
	setupEventSubscriptions(ctx, bus, stateMachine, log)

	// Initialize API router
	router := api.NewRouter(
		cfg,
		log,
		projectRepo,
		serviceRepo,
		nil, // userRepo - implement as needed
		bus,
		coolifyAdapter,
	)
	engine := router.Setup()

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.GetAddress(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("address", srv.Addr).Msg("Starting HTTP server")
		if cfg.Server.TLSEnabled {
			if err := srv.ListenAndServeTLS(cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Failed to start HTTPS server")
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Failed to start HTTP server")
			}
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	cancel() // Cancel root context to stop background goroutines

	log.Info().Msg("Server stopped")
}

// setupEventSubscriptions sets up event subscriptions for workflow processing
func setupEventSubscriptions(ctx context.Context, bus *eventbus.NATSEventBus, sm *workflow.StateMachine, log *logger.Logger) {
	// Subscribe to build events
	bus.Subscribe(ctx, "build.>", func(event *domain.Event) error {
		log.Debug().Str("type", event.Type).Interface("data", event.Data).Msg("Received build event")
		// Process build events and update workflow state
		return nil
	})

	// Subscribe to deployment events
	bus.Subscribe(ctx, "deploy.>", func(event *domain.Event) error {
		log.Debug().Str("type", event.Type).Interface("data", event.Data).Msg("Received deploy event")
		// Process deploy events and update workflow state
		return nil
	})

	// Subscribe to webhook events
	bus.Subscribe(ctx, "webhook.>", func(event *domain.Event) error {
		log.Debug().Str("type", event.Type).Interface("data", event.Data).Msg("Received webhook event")
		// Process webhook events
		return nil
	})
}

// Suppress unused import warning
var _ = rancherAdapter
