package grpc

import (
	"context"
	"os"

	pb "github.com/Oidiral/shipment-customer-system/api/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CustomerClient struct {
	client pb.CustomerServiceClient
	conn   *grpc.ClientConn
}

func NewCustomerClient() (*CustomerClient, error) {
	target := os.Getenv("CUSTOMER_SERVICE_TARGET")
	if target == "" {
		target = "envoy:9090"
	}

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}

	return &CustomerClient{
		client: pb.NewCustomerServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *CustomerClient) Close() error {
	return c.conn.Close()
}

func (c *CustomerClient) UpsertCustomer(ctx context.Context, idn string) (string, error) {
	tracer := otel.Tracer("shipment-service")
	ctx, span := tracer.Start(ctx, "GRPC.UpsertCustomer")
	defer span.End()

	resp, err := c.client.UpsertCustomer(ctx, &pb.UpsertCustomerRequest{Idn: idn})
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	return resp.GetId(), nil
}

func (c *CustomerClient) GetCustomer(ctx context.Context, idn string) (*pb.CustomerResponse, error) {
	tracer := otel.Tracer("shipment-service")
	ctx, span := tracer.Start(ctx, "CustomerClient.GetCustomer")
	defer span.End()

	resp, err := c.client.GetCustomer(ctx, &pb.GetCustomerRequest{
		Idn: idn,
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return resp, nil
}
