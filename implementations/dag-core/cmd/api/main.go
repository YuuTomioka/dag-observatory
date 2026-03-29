// @title DAG Observatory API
// @version 0.1.0
// @description DAG runtime HTTP API (auto-generated via swag).
// @BasePath /
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
		if app.MarketData != nil && app.MarketData.Client != nil {
			app.MarketData.Client.Close()
		}
		consumerCancel()
	}()

	if app.DAGRuntime != nil && app.DAGRuntime.EventStream != nil {
		stream, err := app.DAGRuntime.EventStream.Subscribe(consumerCtx)
		if err != nil {
			log.Fatalf("event stream subscribe failed: %v", err)
		}
		driverStream := make(chan driver.StreamEvent, 32)
		go func() {
			defer close(driverStream)
			for item := range stream {
				select {
				case driverStream <- driver.StreamEvent{Ctx: item.Ctx, Event: item.Event}:
				case <-consumerCtx.Done():
					return
				}
			}
		}()
		go func() {
			if err := app.DAGRuntime.Driver.RunWithContext(consumerCtx, driverStream); err != nil && consumerCtx.Err() == nil {
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
	go func() {
		if app.GRPC != nil {
			log.Printf("grpc server placeholder started on :%s", app.Config.GRPCPort)
		}
	}()
	go func() {
		if app.GraphQL != nil {
			log.Printf("graphql server placeholder started on :%s", app.Config.GraphQLPort)
		}
	}()
	go func() {
		if app.WS != nil {
			log.Printf("ws server placeholder started on :%s", app.Config.WSPort)
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
