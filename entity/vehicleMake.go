package entity

import "time"

type VehicleMake struct {
	ID            int       `json:"id"`
	MakeName      string    `json:"make_name"`
	CountryOrigin string    `json:"country_origin"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}
