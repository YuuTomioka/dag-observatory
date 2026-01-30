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
	app, err := di.NewAppContainer()
	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = app.OTel.Shutdown(ctx)
	}()

	// start server
	go func() {
		if err := app.Echo.Start(":" + app.Config.HTTPPort); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = app.Echo.Shutdown(ctx)
}
