package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	appName    = "cfguardian"
	appVersion = "0.1.0"
)

func main() {
	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting GoConfig Guardian",
		"app", appName,
		"version", appVersion,
	)

	// Create context that listens for termination signals
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	// TODO: Load configuration
	// TODO: Initialize database connection
	// TODO: Initialize Raft consensus
	// TODO: Setup repositories
	// TODO: Setup use cases
	// TODO: Setup HTTP handlers
	// TODO: Start HTTP server

	slog.Info("Application initialized successfully")
	slog.Info("Server will start on :8080 (not implemented yet)")

	// Wait for interrupt signal
	<-ctx.Done()

	// Graceful shutdown
	slog.Info("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// TODO: Close database connections
	// TODO: Stop Raft node
	// TODO: Shutdown HTTP server

	<-shutdownCtx.Done()
	slog.Info("Shutdown complete")
}

func displayBanner() {
	banner := `
   ____      ____                    _______                     ___          
  / ___|    / ___|_   _  __ _ _ __ __| (_) __ _ _ __    
 | |   _   | |  _| | | |/ _' | '__/ _' | |/ _' | '_ \   
 | |__| |  | |_| | |_| | (_| | | | (_| | | (_| | | | |  
  \____|   \____|\__,_|\__,_|_|  \__,_|_|\__,_|_| |_|  
                                                         
  GoConfig Guardian - Distributed Configuration Management
  Version: %s
	`
	fmt.Printf(banner, appVersion)
	fmt.Println()
}

func init() {
	displayBanner()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

