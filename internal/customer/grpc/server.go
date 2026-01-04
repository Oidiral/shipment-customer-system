package grpc

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/Oidiral/shipment-customer-system/api/proto"
	"github.com/Oidiral/shipment-customer-system/internal/customer/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedCustomerServiceServer
	svc    *service.Service
	logger *slog.Logger
}

func NewServer(svc *service.Service, logger *slog.Logger) *Server {
	return &Server{
		svc:    svc,
		logger: logger,
	}
}

func (s *Server) UpsertCustomer(ctx context.Context, req *pb.UpsertCustomerRequest) (*pb.CustomerResponse, error) {
	tracer := otel.Tracer("customer-service")
	ctx, span := tracer.Start(ctx, "gRPC.UpsertCustomer")
	defer span.End()

	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	s.logger.Info("UpsertCustomer called", "idn", req.Idn, "trace_id", traceID)

	if req.Idn == "" {
		return nil, status.Error(codes.InvalidArgument, "idn is required")
	}

	customer, err := s.svc.UpsertCustomer(ctx, req.Idn)
	if err != nil {
		s.logger.Error("failed to upsert customer", "error", err, "trace_id", traceID)
		span.RecordError(err)
		return nil, status.Error(codes.Internal, "failed to upsert customer")
	}

	s.logger.Info("Customer upserted", "id", customer.ID, "idn", customer.IDN, "trace_id", traceID)

	return &pb.CustomerResponse{
		Id:        customer.ID,
		Idn:       customer.IDN,
		CreatedAt: customer.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Server) GetCustomer(ctx context.Context, req *pb.GetCustomerRequest) (*pb.CustomerResponse, error) {
	tracer := otel.Tracer("customer-service")
	ctx, span := tracer.Start(ctx, "gRPC.GetCustomer")
	defer span.End()

	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	s.logger.Info("GetCustomer called", "idn", req.Idn, "trace_id", traceID)

	if req.Idn == "" {
		return nil, status.Error(codes.InvalidArgument, "idn is required")
	}

	customer, err := s.svc.GetCustomer(ctx, req.Idn)
	if err != nil {
		s.logger.Error("failed to get customer", "error", err, "trace_id", traceID)
		span.RecordError(err)
		return nil, status.Error(codes.NotFound, "customer not found")
	}

	return &pb.CustomerResponse{
		Id:        customer.ID,
		Idn:       customer.IDN,
		CreatedAt: customer.CreatedAt.Format(time.RFC3339),
	}, nil
}
