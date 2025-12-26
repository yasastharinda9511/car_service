package request

type ShippingStatusEmailRequest struct {
	ToEmail              string  `json:"to_email"`
	CustomerName         string  `json:"customer_name"`
	CarMake              string  `json:"car_make"`
	CarModel             string  `json:"car_model"`
	CarYear              string  `json:"car_year"`
	ChassisNumber        string  `json:"chassis_number"`
	OldStatus            string  `json:"old_status"`
	NewStatus            string  `json:"new_status"`
	ShippingOrderID      string  `json:"shipping_order_id"`
	VesselName           *string `json:"vessel_name,omitempty"`
	VoyageNumber         *string `json:"voyage_number,omitempty"`
	ContainerNumber      *string `json:"container_number,omitempty"`
	BillOfLading         *string `json:"bill_of_lading,omitempty"`
	PortOfLoading        *string `json:"port_of_loading,omitempty"`
	PortOfDischarge      *string `json:"port_of_discharge,omitempty"`
	EstimatedArrivalDate *string `json:"estimated_arrival_date,omitempty"`
	ActualArrivalDate    *string `json:"actual_arrival_date,omitempty"`
	TrackingURL          *string `json:"tracking_url,omitempty"`
	Notes                *string `json:"notes,omitempty"`
	ContactPerson        *string `json:"contact_person,omitempty"`
}
