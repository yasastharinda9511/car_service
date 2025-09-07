package request

type UpdateVehicleRequest struct {
	Code               *int     `json:"code"`
	Make               *string  `json:"make"`
	Model              *string  `json:"model"`
	TrimLevel          *string  `json:"trim_level"`
	YearOfManufacture  *int     `json:"year_of_manufacture"`
	Color              *string  `json:"color"`
	MileageKm          *int     `json:"mileage_km"`
	ChassisID          *string  `json:"chassis_id"`
	ConditionStatus    *string  `json:"condition_status"`
	YearOfRegistration *int     `json:"year_of_registration"`
	LicensePlate       *string  `json:"license_plate"`
	AuctionGrade       *string  `json:"auction_grade"`
	AuctionPrice       *float64 `json:"auction_price"`
	CIFValue           *float64 `json:"cif_value"`
	Currency           *string  `json:"currency"`
	HSCode             *string  `json:"hs_code"`
	InvoiceFOBJPY      *float64 `json:"invoice_fob_jpy"`
	RegistrationNumber *string  `json:"registration_number"`
}
