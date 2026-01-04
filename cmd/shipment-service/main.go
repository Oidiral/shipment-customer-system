package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	grpcclient "github.com/Oidiral/shipment-customer-system/internal/shipment/grpc"
	httphandler "github.com/Oidiral/shipment-customer-system/internal/shipment/http"
	"github.com/Oidiral/shipment-customer-system/internal/shipment/repo"
	"github.com/Oidiral/shipment-customer-system/internal/shipment/service"
	"github.com/Oidiral/shipment-customer-system/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := telemetry.NewLogger("shipment-service")

	shutdown, err := telemetry.InitTracer(ctx, "shipment-service")
	if err != nil {
		logger.Error("failed to init tracer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			logger.Error("failed to shutdown tracer", "error", err)
		}
	}()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/shipments?sslmode=disable"
	}

	var db *sql.DB
	for i := 0; i < 30; i++ {
		db, err = sql.Open("postgres", dbURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		logger.Info("waiting for database", "attempt", i+1)
		time.Sleep(time.Second)
	}
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("connected to database")

	customerClient, err := grpcclient.NewCustomerClient()
	if err != nil {
		logger.Error("failed to create customer client", "error", err)
		os.Exit(1)
	}
	defer customerClient.Close()

	repository := repo.NewRepository(db)
	svc := service.NewService(repository, customerClient)
	handler := httphandler.NewHandler(svc, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/shipments", handler.CreateShipment)
	mux.HandleFunc("GET /api/v1/shipments/{id}", handler.GetShipment)

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: otelhttp.NewHandler(mux, "shipment-service"),
	}

	go func() {
		logger.Info("starting HTTP server", "port", httpPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown server", "error", err)
	}
}
