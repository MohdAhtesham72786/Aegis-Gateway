package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/example/aegis-gateway/internal/gateway"
	"github.com/example/aegis-gateway/internal/policy"
	"github.com/example/aegis-gateway/internal/adapters/payments"
	"github.com/example/aegis-gateway/internal/adapters/files"
	tele "github.com/example/aegis-gateway/pkg/telemetry"
)

func main() {
	log.Println("starting aegis gateway")

	ctx := context.Background()
	// init telemetry
	tracer, provider := tele.InitForConsole()
	defer func() { _ = provider.Shutdown(ctx) }()

	// load policies
	pl, err := policy.NewLoader("./policies")
	if err != nil {
		log.Fatalf("policy loader init: %v", err)
	}
	go pl.Watch()

	// start mock adapters (in-process) for demo
	paymentsSrv := payments.NewServer()
	filesSrv := files.NewServer()
	go paymentsSrv.Start(9001)
	go filesSrv.Start(9002)

	// gateway
	gw := gateway.New(pl, tracer, "http://localhost:9001", "http://localhost:9002")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: gw.Router(),
	}

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		log.Println("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Println("gateway listening :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
