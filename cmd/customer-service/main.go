package main

import (
	"context"
	"database/sql"
	"net"
	"os"
	"os/signal"
	"time"

	_ "github.com/lib/pq"

	pb "github.com/Oidiral/shipment-customer-system/api/proto"
	customergrpc "github.com/Oidiral/shipment-customer-system/internal/customer/grpc"
	"github.com/Oidiral/shipment-customer-system/internal/customer/repo"
	"github.com/Oidiral/shipment-customer-system/internal/customer/service"
	"github.com/Oidiral/shipment-customer-system/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := telemetry.NewLogger("customer-service")

	shutdown, err := telemetry.InitTracer(ctx, "customer-service")
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

	repository := repo.NewRepository(db)
	svc := service.NewService(repository)
	server := customergrpc.NewServer(svc, logger)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	pb.RegisterCustomerServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	go func() {
		logger.Info("starting gRPC server", "port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	logger.Info("shutting down")
	grpcServer.GracefulStop()
}
