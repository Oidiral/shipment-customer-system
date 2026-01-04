package repo

import (
	"context"
	"database/sql"

	"github.com/Oidiral/shipment-customer-system/internal/shipment/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db}
}

func (r *Repository) CreateShipment(ctx context.Context, route string, price float64, customerID string) (*model.Shipment, error) {
	tracer := otel.Tracer("shipment-service")
	ctx, span := tracer.Start(ctx, "Repository.CreateShipment")
	defer span.End()

	span.SetAttributes(attribute.String("db.operation", "insert"))
	span.SetAttributes(attribute.String("shipment.route", route))
	span.SetAttributes(attribute.Float64("shipment.price", price))
	span.SetAttributes(attribute.String("shipment.customer_id", customerID))

	query := `INSERT INTO shipments (route, price, customer_id) VALUES ($1, $2, $3) RETURNING id, route, price, status, customer_id, created_at`

	var s model.Shipment
	err := r.db.QueryRowContext(ctx, query, route, price, customerID).Scan(&s.ID, &s.Route, &s.Price, &s.Status, &s.CustomerID, &s.CreatedAt)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &s, nil
}

func (r *Repository) GetShipment(ctx context.Context, id string) (*model.Shipment, error) {
	tracer := otel.Tracer("shipment-service")
	ctx, span := tracer.Start(ctx, "Repository.GetShipment")
	defer span.End()

	span.SetAttributes(attribute.String("db.operation", "select"))
	span.SetAttributes(attribute.String("shipment.id", id))

	query := `SELECT id, route, price, status, customer_id, created_at FROM shipments WHERE id = $1`

	var s model.Shipment
	err := r.db.QueryRowContext(ctx, query, id).Scan(&s.ID, &s.Route, &s.Price, &s.Status, &s.CustomerID, &s.CreatedAt)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &s, nil
}
