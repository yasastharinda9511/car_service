package repository

import (
	"car_service/database"
	"car_service/dto/request"
	"car_service/entity"
	"car_service/filters"
	"context"
)

type VehicleRepository struct{}

func NewVehicleRepository() *VehicleRepository {
	return &VehicleRepository{}
}

func (s *VehicleRepository) GetAllVehicles(ctx context.Context, exec database.Executor, limit, offset int, filter filters.Filter) ([]entity.VehicleComplete, error) {
	query := `
		SELECT 
			v.id,
			v.code,
			v.make,
			v.model,
			v.trim_level,
			v.year_of_manufacture,
			v.color,
			v.mileage_km,
			v.chassis_id,
			v.condition_status,
			v.auction_grade,
			v.cif_value,
			v.currency,
			v.created_at,
			v.updated_at,

			COALESCE(vs.id, 0) AS vs_id,
			COALESCE(vs.vehicle_id, 0) AS vs_vehicle_id,
			COALESCE(vs.vessel_name, '') AS vessel_name,
			COALESCE(vs.departure_harbour, '') AS departure_harbour,
			COALESCE(vs.shipment_date, '1970-01-01') AS shipment_date,
			COALESCE(vs.arrival_date, '1970-01-01') AS arrival_date,
			COALESCE(vs.clearing_date, '1970-01-01') AS clearing_date,
			COALESCE(vs.shipping_status, 'PROCESSING') AS shipping_status,

			COALESCE(vf.id, 0) AS vf_id,
			COALESCE(vf.vehicle_id, 0) AS vf_vehicle_id,
			COALESCE(vf.total_cost_lkr, 0) AS total_cost_lkr,
			COALESCE(vf.charges_lkr, 0) AS charges_lkr,
			COALESCE(vf.duty_lkr, 0) AS duty_lkr,
			COALESCE(vf.clearing_lkr, 0) AS clearing_lkr,
			COALESCE(vf.other_expenses_lkr, 0) AS other_expenses_lkr,

			COALESCE(vsl.id, 0) AS vsl_id,
			COALESCE(vsl.vehicle_id, 0) AS vsl_vehicle_id,
			COALESCE(vsl.sold_date, '1970-01-01') AS sold_date,
			COALESCE(vsl.revenue, 0) AS revenue,
			COALESCE(vsl.profit, 0) AS profit,
			COALESCE(vsl.sold_to_name, '') AS sold_to_name,
			COALESCE(vsl.sold_to_title, '') AS sold_to_title,
			COALESCE(vsl.contact_number, '') AS contact_number,
			COALESCE(vsl.customer_address, '') AS customer_address,
			COALESCE(vsl.sale_status, 'AVAILABLE') AS sale_status,

			COALESCE(vp.id, 0) AS vp_id,
			COALESCE(vp.vehicle_id, 0) AS vp_vehicle_id,
			COALESCE(vp.bought_from_name, '') AS bought_from_name,
			COALESCE(vp.bought_from_title, '') AS bought_from_title,
			COALESCE(vp.bought_from_contact, '') AS bought_from_contact,
			COALESCE(vp.bought_from_address, '') AS bought_from_address,
			COALESCE(vp.bought_from_other_contacts, '') AS bought_from_other_contacts,
			COALESCE(vp.purchase_remarks, '') AS purchase_remarks,
			COALESCE(vp.lc_bank, '') AS lc_bank,
			COALESCE(vp.lc_number, '') AS lc_number,
			COALESCE(vp.lc_cost_jpy, 0) AS lc_cost_jpy,
			COALESCE(vp.purchase_date, '1970-01-01') AS purchase_date,

			COALESCE(vi.id, 0) AS vi_id,
			COALESCE(vi.vehicle_id, 0) AS vi_vehicle_id,
			COALESCE(vi.filename, '') AS filename,
			COALESCE(vi.original_name, '') AS original_name,
			COALESCE(vi.file_path, '') AS file_path,
			COALESCE(vi.file_size, 0) AS file_size,
			COALESCE(vi.mime_type, '') AS mime_type,
			COALESCE(vi.is_primary, false) AS is_primary,
			COALESCE(vi.upload_date, '1970-01-01') AS upload_date,
			COALESCE(vi.display_order, 0) AS display_order

		FROM cars.vehicles v
		LEFT JOIN cars.vehicle_shipping vs ON v.id = vs.vehicle_id
		LEFT JOIN cars.vehicle_financials vf ON v.id = vf.vehicle_id
		LEFT JOIN cars.vehicle_sales vsl ON v.id = vsl.vehicle_id
		LEFT JOIN cars.vehicle_purchases vp ON v.id = vp.vehicle_id
		LEFT JOIN cars.vehicle_images vi ON v.id = vi.vehicle_id
	`

	query, args := filter.GetQuery(query, "", "", limit, offset)
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vehicleMap := make(map[int]*entity.VehicleComplete)

	for rows.Next() {
		var vc entity.VehicleComplete
		var img entity.VehicleImage

		err := rows.Scan(
			&vc.Vehicle.ID, &vc.Vehicle.Code, &vc.Vehicle.Make, &vc.Vehicle.Model, &vc.Vehicle.TrimLevel, &vc.Vehicle.YearOfManufacture,
			&vc.Vehicle.Color, &vc.Vehicle.MileageKm, &vc.Vehicle.ChassisID, &vc.Vehicle.ConditionStatus, &vc.Vehicle.AuctionGrade,
			&vc.Vehicle.CIFValue, &vc.Vehicle.Currency, &vc.Vehicle.CreatedAt, &vc.Vehicle.UpdatedAt,

			&vc.VehicleShipping.ID, &vc.VehicleShipping.VehicleID, &vc.VehicleShipping.VesselName, &vc.VehicleShipping.DepartureHarbour, &vc.VehicleShipping.ShipmentDate, &vc.VehicleShipping.ArrivalDate,
			&vc.VehicleShipping.ClearingDate, &vc.VehicleShipping.ShippingStatus,

			&vc.VehicleFinancials.ID, &vc.VehicleFinancials.VehicleID, &vc.VehicleFinancials.TotalCostLKR, &vc.VehicleFinancials.ChargesLKR, &vc.VehicleFinancials.DutyLKR, &vc.VehicleFinancials.ClearingLKR, &vc.VehicleFinancials.OtherExpensesLKR,

			&vc.VehicleSales.ID, &vc.VehicleSales.VehicleID, &vc.VehicleSales.SoldDate, &vc.VehicleSales.Revenue, &vc.VehicleSales.Profit, &vc.VehicleSales.SoldToName, &vc.VehicleSales.SoldToTitle,
			&vc.VehicleSales.ContactNumber, &vc.VehicleSales.CustomerAddress, &vc.VehicleSales.SaleStatus,

			&vc.VehiclePurchase.ID, &vc.VehiclePurchase.VehicleID, &vc.VehiclePurchase.BoughtFromName, &vc.VehiclePurchase.BoughtFromTitle,
			&vc.VehiclePurchase.BoughtFromContact, &vc.VehiclePurchase.BoughtFromAddress, &vc.VehiclePurchase.BoughtFromOtherContacts,
			&vc.VehiclePurchase.PurchaseRemarks, &vc.VehiclePurchase.LCBank, &vc.VehiclePurchase.LCNumber, &vc.VehiclePurchase.LCCostJPY, &vc.VehiclePurchase.PurchaseDate,

			&img.ID, &img.VehicleID, &img.Filename, &img.OriginalName,
			&img.FilePath, &img.FileSize, &img.MimeType, &img.IsPrimary,
			&img.UploadDate, &img.DisplayOrder,
		)
		if err != nil {
			return nil, err
		}

		vehicleID := vc.Vehicle.ID

		if existingVehicle, exists := vehicleMap[int(vehicleID)]; exists {
			if img.ID > 0 {
				existingVehicle.VehicleImages = append(existingVehicle.VehicleImages, img)
			}
		} else {
			vc.VehicleImages = []entity.VehicleImage{}
			if img.ID > 0 {
				vc.VehicleImages = append(vc.VehicleImages, img)
			}
			vehicleMap[int(vehicleID)] = &vc
		}
	}

	// Convert map to slice
	var vehicles []entity.VehicleComplete
	for _, vehicle := range vehicleMap {
		vehicles = append(vehicles, *vehicle)
	}

	return vehicles, nil
}

func (s *VehicleRepository) GetAllVehicleCount(ctx context.Context, exec database.Executor, filter filters.Filter) (int64, error) {
	var count int64
	query := `SELECT COUNT(*)
        FROM cars.vehicles v
        LEFT JOIN cars.vehicle_shipping vs ON v.id = vs.vehicle_id
        LEFT JOIN cars.vehicle_financials vf ON v.id = vf.vehicle_id
        LEFT JOIN cars.vehicle_sales vsl ON v.id = vsl.vehicle_id
        LEFT JOIN cars.vehicle_purchases vp ON v.id = vp.vehicle_id`

	query, args := filter.GetQuery(query, "", "", -1, -1)

	err := exec.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err

	}
	return count, nil
}

func (s *VehicleRepository) GetVehicleByID(ctx context.Context, exec database.Executor, id int64) (*entity.Vehicle, error) {
	query := `SELECT
        v.id,
		v.code,
		v.make,
		v.model,
		v.trim_level,
		v.year_of_manufacture,
		v.color,
		v.mileage_km,
		v.chassis_id,
		v.condition_status,
		v.auction_grade,
		v.cif_value,
		v.currency,
		v.created_at,
		v.updated_at
		FROM cars.vehicles v
		WHERE v.id = $1`

	var vehicle entity.Vehicle
	err := exec.QueryRowContext(ctx, query, id).Scan(&vehicle.ID, &vehicle.Code, &vehicle.Make, &vehicle.Model, &vehicle.TrimLevel, &vehicle.YearOfManufacture,
		&vehicle.Color, &vehicle.MileageKm, &vehicle.ChassisID, &vehicle.ConditionStatus, &vehicle.AuctionGrade,
		&vehicle.CIFValue, &vehicle.Currency, &vehicle.CreatedAt, &vehicle.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &vehicle, nil
}

func (s *VehicleRepository) Insert(ctx context.Context, exec database.Executor, req request.CreateVehicleRequest) (int64, error) {
	var vehicleID int64
	query := `
        INSERT INTO cars.vehicles (code, make, model, trim_level, year_of_manufacture, color,
            mileage_km, chassis_id, condition_status, auction_grade, cif_value, currency)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING id
    `
	err := exec.QueryRowContext(ctx, query,
		req.Code, req.Make, req.Model, req.TrimLevel,
		req.YearOfManufacture, req.Color, req.MileageKm, req.ChassisID,
		req.ConditionStatus, req.AuctionGrade, req.CIFValue, req.Currency,
	).Scan(&vehicleID)
	return vehicleID, err
}

func (s *VehicleRepository) UpdateVehicleDetails(ctx context.Context, exec database.Executor, vehicleID int64, req *request.UpdateVehicleRequest) error {
	query := `
       UPDATE cars.vehicles
       SET code = COALESCE($2, code),
           make = COALESCE($3, make),
           model = COALESCE($4, model),
           trim_level = COALESCE($5, trim_level),
           year_of_manufacture = COALESCE($6, year_of_manufacture),
           color = COALESCE($7, color),
           mileage_km = COALESCE($8, mileage_km),
           chassis_id = COALESCE($9, chassis_id),
           condition_status = COALESCE($10, condition_status),
           year_of_registration = COALESCE($11, year_of_registration),
           license_plate = COALESCE($12, license_plate),
           auction_grade = COALESCE($13, auction_grade),
           auction_price = COALESCE($14, auction_price),
           cif_value = COALESCE($15, cif_value),
           currency = COALESCE($16, currency),
           hs_code = COALESCE($17, hs_code),
           invoice_fob_jpy = COALESCE($18, invoice_fob_jpy),
           registration_number = COALESCE($19, registration_number),
           record_date = COALESCE($20, record_date),
           updated_at = CURRENT_TIMESTAMP
       WHERE id = $1
   `

	_, err := exec.QueryContext(ctx, query, vehicleID, req.Code, req.Make, req.Model, req.TrimLevel,
		req.YearOfManufacture, req.Color, req.MileageKm, req.ChassisID, req.ConditionStatus,
		req.YearOfRegistration, req.LicensePlate, req.AuctionGrade, req.AuctionPrice, req.CIFValue,
		req.Currency, req.HSCode, req.InvoiceFOBJPY, req.RegistrationNumber, req.RecordDate)

	return err

}
func (s *VehicleRepository) GetVehicleBrandCount(ctx context.Context, exec database.Executor, filter filters.Filter) (map[string]int, error) {
	query := `SELECT
	make,
    COUNT(*) as vehicle_count
	FROM cars.vehicles v`

	query, args := filter.GetQuery(query, "v.make", "", -1, -1)
	rows, err := exec.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]int)

	for rows.Next() {
		var status string
		var count int

		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, err
		}

		results[status] = count
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
