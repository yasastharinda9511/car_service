package entity

import "time"

type VehicleModel struct {
	ID               int       `json:"id"`
	MakeID           int       `json:"make_id"`
	ModelName        string    `json:"model_name"`
	BodyType         *string   `json:"body_type"`
	FuelType         *string   `json:"fuel_type"`
	TransmissionType *string   `json:"transmission_type"`
	EngineSizeCC     *int      `json:"engine_size_cc"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
}
