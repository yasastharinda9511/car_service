package request

type CreateVehicleMake struct {
	MakeName      *string `json:"make_name"`      // Required
	CountryOrigin *string `json:"country_origin"` // Optional, defaults to "Japan"
	LogoURL       *string `json:"logo_url"`       // Optional, S3 path to logo
	IsActive      *bool   `json:"is_active"`      // Optional, defaults to true
}
