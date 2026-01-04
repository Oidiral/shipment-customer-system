package service

import (
	"context"
	"errors"
	"regexp"
	"time"

	grpcclient "github.com/Oidiral/shipment-customer-system/internal/shipment/grpc"
	"github.com/Oidiral/shipment-customer-system/internal/shipment/repo"
	"go.opentelemetry.io/otel"
)

var idnRegex = regexp.MustCompile(`^\d{12}$`)

var ErrInvalidIDN = errors.New("invalid IDN: must be exactly 12 digits")

type Shipment struct {
	ID         string
	Route      string
	Price      float64
	Status     string
	CustomerID string
	CreatedAt  time.Time
}

type CreateShipmentRequest struct {
	Route string
	Price float64
	IDN   string
}

type Service struct {
	repo           *repo.Repository
	customerClient *grpcclient.CustomerClient
}

func NewService(repo *repo.Repository, customerClient *grpcclient.CustomerClient) *Service {
	return &Service{
		repo:           repo,
		customerClient: customerClient,
	}
}

func (s *Service) CreateShipment(ctx context.Context, req CreateShipmentRequest) (*Shipment, error) {
	tracer := otel.Tracer("shipment-service")
	ctx, span := tracer.Start(ctx, "Service.CreateShipment")
	defer span.End()

	if !idnRegex.MatchString(req.IDN) {
		return nil, ErrInvalidIDN
	}

	customerID, err := s.customerClient.UpsertCustomer(ctx, req.IDN)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	shipment, err := s.repo.CreateShipment(ctx, req.Route, req.Price, customerID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &Shipment{
		ID:         shipment.ID,
		Route:      shipment.Route,
		Price:      shipment.Price,
		Status:     shipment.Status,
		CustomerID: shipment.CustomerID,
		CreatedAt:  shipment.CreatedAt,
	}, nil
}

func (s *Service) GetShipment(ctx context.Context, id string) (*Shipment, error) {
	tracer := otel.Tracer("shipment-service")
	ctx, span := tracer.Start(ctx, "Service.GetShipment")
	defer span.End()

	shipment, err := s.repo.GetShipment(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &Shipment{
		ID:         shipment.ID,
		Route:      shipment.Route,
		Price:      shipment.Price,
		Status:     shipment.Status,
		CustomerID: shipment.CustomerID,
		CreatedAt:  shipment.CreatedAt,
	}, nil
}
