package model

import "time"

type Customer struct {
	ID        string
	IDN       string
	CreatedAt time.Time
}