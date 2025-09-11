package response

import "car_service/entity"

type VehiclesResponse struct {
	Vehicles []entity.VehicleComplete `json:"data"`
	Meta     Meta                     `json:"meta"`
}
