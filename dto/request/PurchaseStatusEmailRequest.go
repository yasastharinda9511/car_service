package request

type PurchaseStatusEmailRequest struct {
	ToEmail              string  `json:"to_email"`
	CustomerName         string  `json:"customer_name"`
	CarMake              string  `json:"car_make"`
	CarModel             string  `json:"car_model"`
	CarYear              string  `json:"car_year"`
	ChassisNumber        string  `json:"chassis_number"`
	OldStatus            string  `json:"old_status"`
	NewStatus            string  `json:"new_status"`
	PurchaseOrderID      string  `json:"purchase_order_id"`
	LCNumber             *string `json:"lc_number,omitempty"`
	SupplierName         *string `json:"supplier_name,omitempty"`
	PortOfLoading        *string `json:"port_of_loading,omitempty"`
	ExpectedShippingDate *string `json:"expected_shipping_date,omitempty"`
	PurchasePrice        *string `json:"purchase_price,omitempty"`
	Currency             *string `json:"currency,omitempty"`
	Notes                *string `json:"notes,omitempty"`
	ContactPerson        *string `json:"contact_person,omitempty"`
}
