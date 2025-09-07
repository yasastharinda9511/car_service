package entity

import "time"

type CustomerOrder struct {
	ID                   int64                  `json:"id" database:"id"`
	OrderNumber          string                 `json:"order_number" database:"order_number"`
	CustomerID           *int64                 `json:"customer_id" database:"customer_id"`
	PreferredMake        *string                `json:"preferred_make" database:"preferred_make"`
	PreferredModel       *string                `json:"preferred_model" database:"preferred_model"`
	PreferredYearMin     *int                   `json:"preferred_year_min" database:"preferred_year_min"`
	PreferredYearMax     *int                   `json:"preferred_year_max" database:"preferred_year_max"`
	PreferredColor       *string                `json:"preferred_color" database:"preferred_color"`
	PreferredTrimLevel   *string                `json:"preferred_trim_level" database:"preferred_trim_level"`
	MaxMileageKm         *int                   `json:"max_mileage_km" database:"max_mileage_km"`
	MinAuctionGrade      *string                `json:"min_auction_grade" database:"min_auction_grade"`
	RequiredFeatures     map[string]interface{} `json:"required_features" database:"required_features"`
	OrderType            string                 `json:"order_type" database:"order_type"`
	ExpectedDeliveryDate *time.Time             `json:"expected_delivery_date" database:"expected_delivery_date"`
	PriorityLevel        string                 `json:"priority_level" database:"priority_level"`
	PreferredPort        *string                `json:"preferred_port" database:"preferred_port"`
	ShippingMethod       string                 `json:"shipping_method" database:"shipping_method"`
	IncludeInsurance     bool                   `json:"include_insurance" database:"include_insurance"`
	BudgetMin            *float64               `json:"budget_min" database:"budget_min"`
	BudgetMax            *float64               `json:"budget_max" database:"budget_max"`
	PaymentMethod        string                 `json:"payment_method" database:"payment_method"`
	DownPayment          *float64               `json:"down_payment" database:"down_payment"`
	SpecialRequests      *string                `json:"special_requests" database:"special_requests"`
	InternalNotes        *string                `json:"internal_notes" database:"internal_notes"`
	OrderStatus          string                 `json:"order_status" database:"order_status"`
	IsDraft              bool                   `json:"is_draft" database:"is_draft"`
	OrderDate            time.Time              `json:"order_date" database:"order_date"`
	CompletedDate        *time.Time             `json:"completed_date" database:"completed_date"`
	CreatedAt            time.Time              `json:"created_at" database:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at" database:"updated_at"`
}
