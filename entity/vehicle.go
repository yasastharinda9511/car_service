package entity

import "time"

type Vehicle struct {
	ID                 int64      `json:"id" database:"id"`
	Code               string     `json:"code" database:"code"`
	Make               string     `json:"make" database:"make"`
	MakeID             *int       `json:"make_id" database:"make_id"`
	Model              string     `json:"model" database:"model"`
	TrimLevel          *string    `json:"trim_level" database:"trim_level"`
	YearOfManufacture  int        `json:"year_of_manufacture" database:"year_of_manufacture"`
	Color              string     `json:"color" database:"color"`
	MileageKm          *int       `json:"mileage_km" database:"mileage_km"`
	ChassisID          string     `json:"chassis_id" database:"chassis_id"`
	ConditionStatus    string     `json:"condition_status" database:"condition_status"`
	YearOfRegistration *int       `json:"year_of_registration" database:"year_of_registration"`
	LicensePlate       *string    `json:"license_plate" database:"license_plate"`
	AuctionGrade       *string    `json:"auction_grade" database:"auction_grade"`
	AuctionPrice       *float64   `json:"auction_price" database:"auction_price"`
	CIFValue           *float64   `json:"cif_value" database:"cif_value"`
	Currency           string     `json:"currency" database:"currency"`
	HSCode             *string    `json:"hs_code" database:"hs_code"`
	InvoiceFOBJPY      *float64   `json:"invoice_fob_jpy" database:"invoice_fob_jpy"`
	RegistrationNumber *string    `json:"registration_number" database:"registration_number"`
	RecordDate         *time.Time `json:"record_date" database:"record_date"`
	IsFeatured         bool       `json:"is_featured" database:"is_featured"`
	FeaturedAt         *time.Time `json:"featured_at" database:"featured_at"`
	CreatedAt          time.Time  `json:"created_at" database:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" database:"updated_at"`
}
