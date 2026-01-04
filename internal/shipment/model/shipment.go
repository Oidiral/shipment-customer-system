package model

import "time"

type Shipment struct {
	ID         string
	Route      string
	Price      float64
	Status     string
	CustomerID string
	CreatedAt  time.Time
}
