package response

import "time"

// PublicVehicleResponse contains vehicle data that can be shared publicly
type PublicVehicleResponse struct {
	// Basic vehicle information (always included)
	Code              string `json:"code"`
	Make              string `json:"make"`
	Model             string `json:"model"`
	YearOfManufacture int    `json:"year_of_manufacture"`
	Color             string `json:"color"`
	ChassisID         string `json:"chassis_id"`
	ConditionStatus   string `json:"condition_status"`

	// Optional fields based on include_details
	TrimLevel          *string  `json:"trim_level,omitempty"`
	MileageKm          *int     `json:"mileage_km,omitempty"`
	YearOfRegistration *int     `json:"year_of_registration,omitempty"`
	AuctionGrade       *string  `json:"auction_grade,omitempty"`
	AuctionPrice       *float64 `json:"auction_price,omitempty"`
	Currency           string   `json:"currency,omitempty"`

	// Shipping details (if included)
	ShippingStatus   *string    `json:"shipping_status,omitempty"`
	VesselName       *string    `json:"vessel_name,omitempty"`
	DepartureHarbour *string    `json:"departure_harbour,omitempty"`
	ShipmentDate     *time.Time `json:"shipment_date,omitempty"`
	ArrivalDate      *time.Time `json:"arrival_date,omitempty"`
	ClearingDate     *time.Time `json:"clearing_date,omitempty"`

	// Financial details (if included)
	TotalCostJPY *float64 `json:"total_cost_jpy,omitempty"`

	// Purchase details (if included)
	PurchaseStatus *string    `json:"purchase_status,omitempty"`
	PurchaseDate   *time.Time `json:"purchase_date,omitempty"`

	// Images (if included)
	Images []VehicleImageResponse `json:"images,omitempty"`

	// Token info
	ShareTokenExpiresAt time.Time `json:"share_token_expires_at"`
}

type VehicleImageResponse struct {
	ID        int    `json:"id"`
	ImageURL  string `json:"image_url"`
	IsPrimary bool   `json:"is_primary"`
}
