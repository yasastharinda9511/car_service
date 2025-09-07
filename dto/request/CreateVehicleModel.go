package request

type CreateVehicleModel struct {
	MakeID           int     `json:"make_id"`           // Required
	ModelName        *string `json:"model_name"`        // Required
	BodyType         *string `json:"body_type"`         // Optional
	FuelType         *string `json:"fuel_type"`         // Optional
	TransmissionType *string `json:"transmission_type"` // Optional
	EngineSizeCC     *int    `json:"engine_size_cc"`    // Optional
	IsActive         *bool   `json:"is_active"`         // Optional, defaults to true
}
