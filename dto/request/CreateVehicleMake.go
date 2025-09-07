package request

type CreateVehicleMake struct {
	MakeName      *string `json:"make_name"`      // Required
	CountryOrigin *string `json:"country_origin"` // Optional, defaults to "Japan"
	IsActive      *bool   `json:"is_active"`      // Optional, defaults to true
}
