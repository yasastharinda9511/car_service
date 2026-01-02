package entity

import "time"

type VehicleShareToken struct {
	ID             int64     `json:"id" database:"id"`
	VehicleID      int64     `json:"vehicle_id" database:"vehicle_id"`
	Token          string    `json:"token" database:"token"`
	ExpiresAt      time.Time `json:"expires_at" database:"expires_at"`
	IncludeDetails []string  `json:"include_details" database:"include_details"`
	CreatedBy      string    `json:"created_by" database:"created_by"`
	CreatedAt      time.Time `json:"created_at" database:"created_at"`
	IsActive       bool      `json:"is_active" database:"is_active"`
}
