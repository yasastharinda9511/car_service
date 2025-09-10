package entity

type VehicleComplete struct {
	Vehicle           Vehicle           `json:"vehicle"`
	VehiclePurchase   VehiclePurchase   `json:"vehicle_purchase"`
	VehicleShipping   VehicleShipping   `json:"vehicle_shipping"`
	VehicleFinancials VehicleFinancials `json:"vehicle_financials"`
	VehicleSales      VehicleSales      `json:"vehicle_sales"`
	VehicleImage      VehicleImage      `json:"vehicle_image"`
}
