package entity

import "time"

type VehicleMake struct {
	ID            int       `json:"id"`
	MakeName      string    `json:"make_name"`
	CountryOrigin string    `json:"country_origin"`
	LogoURL       *string   `json:"logo_url"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}
