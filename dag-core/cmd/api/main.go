package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dag-observatory/dag-core/internal/di"
	"dag-observatory/dag-core/internal/domain/dagruntime/driver"
)

func main() {
	app, err := di.NewAppContainer()
	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = app.OTel.Shutdown(ctx)
		if app.DAGRuntime != nil && app.DAGRuntime.StateStore != nil && app.DAGRuntime.StateStore.Close != nil {
			_ = app.DAGRuntime.StateStore.Close()
		}
		if app.DAGRuntime != nil && app.DAGRuntime.EventProducer != nil {
			_ = app.DAGRuntime.EventProducer.Close()
		}
		if app.DAGRuntime != nil && app.DAGRuntime.EventConsumer != nil {
			_ = app.DAGRuntime.EventConsumer.Close()
		}
		consumerCancel()
	}()

	if app.DAGRuntime != nil && app.DAGRuntime.EventConsumer != nil {
		stream := make(chan driver.StreamEvent, 32)
		go func() {
			if err := app.DAGRuntime.EventConsumer.RunWithContext(consumerCtx, stream); err != nil && consumerCtx.Err() == nil {
				log.Printf("kafka consumer stopped: %v", err)
			}
		}()
		go func() {
			if err := app.DAGRuntime.Driver.RunWithContext(consumerCtx, stream); err != nil && consumerCtx.Err() == nil {
				log.Printf("dag driver stopped: %v", err)
			}
		}()
	}

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
