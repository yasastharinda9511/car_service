package request

type ShippingDetailsRequest struct {
	VesselName       *string `json:"vessel_name"`
	DepartureHarbour *string `json:"departure_harbour"`
	ShipmentDate     *string `json:"shipment_date"`   // "2024-01-15T10:30:00Z"
	ArrivalDate      *string `json:"arrival_date"`    // "2024-01-30T14:20:00Z"
	ClearingDate     *string `json:"clearing_date"`   // "2024-02-05T09:15:00Z"
	ShippingStatus   string  `json:"shipping_status"` // Required field
}
