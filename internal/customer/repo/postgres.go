package repo

import (
	"context"
	"database/sql"

	"github.com/Oidiral/shipment-customer-system/internal/customer/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db}
}

func (r *Repository) UpsertCustomer(ctx context.Context, idn string) (*model.Customer, error) {
	tracer := otel.Tracer("customer-service")
	ctx, span := tracer.Start(ctx, "Repo.UpsertCustomer")
	defer span.End()

	span.SetAttributes(attribute.String("db.operation", "upsert"))
	span.SetAttributes(attribute.String("customer.idn", idn))

	query := `
		INSERT INTO customers (idn)
		VALUES ($1)
		ON CONFLICT (idn) DO UPDATE SET idn = EXCLUDED.idn
		RETURNING id, idn, created_at
	`

	var c model.Customer
	err := r.db.QueryRowContext(ctx, query, idn).Scan(&c.ID, &c.IDN, &c.CreatedAt)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &c, nil
}

func (r *Repository) GetCustomerByIDN(ctx context.Context, idn string) (*model.Customer, error) {
	tracer := otel.Tracer("customer-service")
	ctx, span := tracer.Start(ctx, "Repository.GetCustomerByIDN")
	defer span.End()

	span.SetAttributes(attribute.String("db.operation", "select"))
	span.SetAttributes(attribute.String("customer.idn", idn))

	query := `SELECT id, idn, created_at FROM customers WHERE idn = $1`

	var c model.Customer
	err := r.db.QueryRowContext(ctx, query, idn).Scan(&c.ID, &c.IDN, &c.CreatedAt)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &c, nil
}
