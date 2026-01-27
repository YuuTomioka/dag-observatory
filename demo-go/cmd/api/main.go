package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dag-observatory/demo-go/internal/di"
)

func main() {
	cfg := di.NewConfig()

	otelc, err := di.NewOTelContainer(cfg)
	if err != nil {
		log.Fatalf("otel init failed: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = otelc.Shutdown(ctx)
	}()

	e, err := di.NewEchoContainer(cfg, otelc)
	if err != nil {
		log.Fatalf("echo init failed: %v", err)
	}

	// start server
	go func() {
		if err := e.Start(":" + cfg.HTTPPort); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = e.Shutdown(ctx)
}
