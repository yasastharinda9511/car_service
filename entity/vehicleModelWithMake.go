package entity

type VehicleModelWithMake struct {
	VehicleModel
	MakeName string `json:"make_name"`
}
