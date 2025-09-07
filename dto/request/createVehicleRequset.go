package request

type CreateVehicleRequest struct {
	Code              int      `json:"code" binding:"required"`
	Make              string   `json:"make" binding:"required"`
	Model             string   `json:"model" binding:"required"`
	TrimLevel         *string  `json:"trim_level"`
	YearOfManufacture int      `json:"year_of_manufacture" binding:"required"`
	Color             string   `json:"color" binding:"required"`
	MileageKm         *int     `json:"mileage_km"`
	ChassisID         string   `json:"chassis_id" binding:"required"`
	ConditionStatus   string   `json:"condition_status"`
	AuctionGrade      *string  `json:"auction_grade"`
	CIFValue          *float64 `json:"cif_value"`
	Currency          string   `json:"currency"`
}
