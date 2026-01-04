package service

import (
	"context"

	"github.com/Oidiral/shipment-customer-system/internal/customer/model"
	"github.com/Oidiral/shipment-customer-system/internal/customer/repo"
	"go.opentelemetry.io/otel"
)

type Service struct {
	repo *repo.Repository
}

func NewService(repo *repo.Repository) *Service {
	return &Service{repo}
}

func (s *Service) UpsertCustomer(ctx context.Context, idn string) (*model.Customer, error) {
	tracer := otel.Tracer("customer-service")
	ctx, span := tracer.Start(ctx, "Service.UpsertCustomer")
	defer span.End()

	return s.repo.UpsertCustomer(ctx, idn)
}

func (s *Service) GetCustomer(ctx context.Context, idn string) (*model.Customer, error) {
	tracer := otel.Tracer("customer-service")
	ctx, span := tracer.Start(ctx, "Service.GetCustomer")
	defer span.End()

	return s.repo.GetCustomerByIDN(ctx, idn)
}
