package request

type CreateOrderRequest struct {
	CustomerName     string   `json:"customer_name" binding:"required"`
	CustomerTitle    string   `json:"customer_title"`
	ContactNumber    string   `json:"contact_number" binding:"required"`
	Email            *string  `json:"email"`
	Address          *string  `json:"address"`
	PreferredMake    string   `json:"preferred_make" binding:"required"`
	PreferredModel   string   `json:"preferred_model" binding:"required"`
	PreferredColor   string   `json:"preferred_color" binding:"required"`
	PreferredYear    int      `json:"preferred_year"`
	TrimLevel        *string  `json:"trim_level"`
	MaxMileage       *int     `json:"max_mileage"`
	MinAuctionGrade  *string  `json:"min_auction_grade"`
	RequiredFeatures []string `json:"required_features"`
	OrderType        string   `json:"order_type"`
	ExpectedDelivery *string  `json:"expected_delivery"`
	Priority         string   `json:"priority"`
	PreferredPort    *string  `json:"preferred_port"`
	ShippingMethod   string   `json:"shipping_method"`
	IncludeInsurance bool     `json:"include_insurance"`
	BudgetMin        *float64 `json:"budget_min"`
	BudgetMax        *float64 `json:"budget_max"`
	PaymentMethod    string   `json:"payment_method"`
	DownPayment      *float64 `json:"down_payment"`
	SpecialRequests  *string  `json:"special_requests"`
	InternalNotes    *string  `json:"internal_notes"`
	IsDraft          bool     `json:"is_draft"`
}
