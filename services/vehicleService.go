package services

import (
	"car_service/database"
	"car_service/dto/request"
	"car_service/entity"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type VehicleService struct {
	db *database.Database
}

func NewVehicleService(db *database.Database) *VehicleService {
	return &VehicleService{db: db}
}

func (s *VehicleService) GetAllVehicleCount() (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM vehicles`

	err := s.db.Db.QueryRow(query).Scan(&count)
	if err != nil {
		// Log the error if you have a logger
		// log.Printf("Error getting vehicle count: %v", err)
		return 0, err
	}

	return count, nil
}

func (s *VehicleService) GetAllVehicles(limit, offset int) ([]entity.VehicleComplete, error) {
	query := `
		SELECT 
			v.id, v.code, v.make, v.model, v.trim_level, v.year_of_manufacture, 
			v.color, v.mileage_km, v.chassis_id, v.condition_status, v.auction_grade,
			v.cif_value, v.currency, v.created_at, v.updated_at,
			
			vs.id,
			vs.vehicle_id,
			vs.vessel_name, vs.departure_harbour, vs.shipment_date, vs.arrival_date, 
			vs.clearing_date, vs.shipping_status,
			
			vf.id, vf.vehicle_id,
			vf.total_cost_lkr, vf.charges_lkr, vf.duty_lkr, vf.clearing_lkr, vf.other_expenses_lkr,
			
			vsl.id, vsl.vehicle_id,
			vsl.sold_date, vsl.revenue, vsl.profit, vsl.sold_to_name, vsl.sold_to_title,
			vsl.contact_number, vsl.customer_address, vsl.sale_status,
			
			vp.id, vp.vehicle_id,
		    vp.bought_from_name,
		    vp.bought_from_title,
		    vp.bought_from_contact,
		    vp.bought_from_address,
		    vp.bought_from_other_contacts,
		    vp.purchase_remarks,
		    vp.lc_bank,
		    vp.lc_number,
		    vp.lc_cost_jpy,
		    vp.purchase_date
		
			
		FROM vehicles v
		LEFT JOIN vehicle_shipping vs ON v.id = vs.vehicle_id
		LEFT JOIN vehicle_financials vf ON v.id = vf.vehicle_id
		LEFT JOIN vehicle_sales vsl ON v.id = vsl.vehicle_id
		LEFT JOIN vehicle_purchases vp ON v.id = vp.vehicle_id
		ORDER BY v.created_at DESC          
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []entity.VehicleComplete
	for rows.Next() {
		var vc entity.VehicleComplete
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
		)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, vc)
	}

	return vehicles, nil
}

func (s *VehicleService) GetVehicleByID(id int64) (*entity.VehicleComplete, error) {
	query := `
		SELECT 
			v.id, v.code, v.make, v.model, v.trim_level, v.year_of_manufacture, 
			v.color, v.mileage_km, v.chassis_id, v.condition_status, v.auction_grade,
			v.cif_value, v.currency, v.created_at, v.updated_at,
			
			vs.id,
			vs.vehicle_id,
			COALESCE(vs.vessel_name, '') as vessel_name, 
			COALESCE(vs.departure_harbour, '') as departure_harbour, 
			vs.shipment_date, vs.arrival_date, vs.clearing_date, 
			COALESCE(vs.shipping_status, 'PROCESSING') as shipping_status,
			
			vf.id, vf.vehicle_id,
			COALESCE(vf.total_cost_lkr, 0) as total_cost_lkr, 
			vf.charges_lkr, vf.duty_lkr, vf.clearing_lkr, vf.other_expenses_lkr,
			
			vsl.id, vsl.vehicle_id,
			vsl.sold_date, vsl.revenue, vsl.profit, vsl.sold_to_name, vsl.sold_to_title,
			vsl.contact_number, vsl.customer_address, 
			COALESCE(vsl.sale_status, 'AVAILABLE') as sale_status,
		
			vp.id, vp.vehicle_id,
		    vp.bought_from_name,
		    vp.bought_from_title,
		    vp.bought_from_contact,
		    vp.bought_from_address,
		    vp.bought_from_other_contacts,
		    vp.purchase_remarks,
		    vp.lc_bank,
		    vp.lc_number,
		    vp.lc_cost_jpy,
		    vp.purchase_date
		    
		FROM vehicles v
		LEFT JOIN vehicle_shipping vs ON v.id = vs.vehicle_id
		LEFT JOIN vehicle_financials vf ON v.id = vf.vehicle_id
		LEFT JOIN vehicle_sales vsl ON v.id = vsl.vehicle_id
		LEFT JOIN vehicle_purchases vp ON v.id = vp.vehicle_id
		WHERE v.id = $1
	`

	var vc entity.VehicleComplete
	err := s.db.Db.QueryRow(query, id).Scan(
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
	)
	if err != nil {
		return nil, err
	}

	return &vc, nil
}

func (s *VehicleService) CreateVehicle(req request.CreateVehicleRequest) (*entity.Vehicle, error) {
	// Start transaction
	tx, err := s.db.Db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Will be ignored if tx is committed

	// Insert vehicle
	var vehicleID int64
	vehicleQuery := `
        INSERT INTO vehicles (code, make, model, trim_level, year_of_manufacture, color, 
            mileage_km, chassis_id, condition_status, auction_grade, cif_value, currency)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING id
    `

	err = tx.QueryRow(vehicleQuery, req.Code, req.Make, req.Model, req.TrimLevel,
		req.YearOfManufacture, req.Color, req.MileageKm, req.ChassisID,
		req.ConditionStatus, req.AuctionGrade, req.CIFValue, req.Currency).Scan(&vehicleID)
	if err != nil {
		return nil, err
	}

	// Insert default shipping record
	_, err = tx.Exec(`
        INSERT INTO vehicle_shipping (vehicle_id, shipping_status) 
        VALUES ($1, 'PROCESSING')
    `, vehicleID)
	if err != nil {
		return nil, err
	}

	// Insert default financial record with zero costs
	_, err = tx.Exec(`
        INSERT INTO vehicle_financials (vehicle_id, charges_lkr, tt_lkr, duty_lkr, 
            clearing_lkr, other_expenses_lkr, total_cost_lkr) 
        VALUES ($1, 0, 0, 0, 0, 0, 0)
    `, vehicleID)
	if err != nil {
		return nil, err
	}

	// Insert default sales record
	_, err = tx.Exec(`
        INSERT INTO vehicle_sales (vehicle_id, sale_status) 
        VALUES ($1, 'AVAILABLE')
    `, vehicleID)
	if err != nil {
		return nil, err
	}

	// Insert default purchase record (can be updated later)
	_, err = tx.Exec(`
        INSERT INTO vehicle_purchases (vehicle_id) 
        VALUES ($1)
    `, vehicleID)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// Fetch the complete vehicle record
	vehicle := &entity.Vehicle{}
	selectQuery := `
        SELECT id, code, make, model, trim_level, year_of_manufacture, color, 
            mileage_km, chassis_id, condition_status, auction_grade, cif_value, currency, 
            created_at, updated_at 
        FROM vehicles WHERE id = $1
    `

	err = s.db.Db.QueryRow(selectQuery, vehicleID).Scan(
		&vehicle.ID, &vehicle.Code, &vehicle.Make, &vehicle.Model, &vehicle.TrimLevel,
		&vehicle.YearOfManufacture, &vehicle.Color, &vehicle.MileageKm, &vehicle.ChassisID,
		&vehicle.ConditionStatus, &vehicle.AuctionGrade, &vehicle.CIFValue, &vehicle.Currency,
		&vehicle.CreatedAt, &vehicle.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return vehicle, nil
}

func (s *VehicleService) UpdateShippingStatus(vehicleID int64, detailsRequest request.ShippingDetailsRequest) error {
	query := `
        UPDATE vehicle_shipping 
        SET vessel_name = $2,
            departure_harbour = $3,
            shipment_date = $4,
            arrival_date = $5,
            clearing_date = $6,
            shipping_status = $7,
            updated_at = CURRENT_TIMESTAMP
        WHERE vehicle_id = $1
    `

	_, err := s.db.Db.Exec(query, vehicleID, detailsRequest.VesselName, detailsRequest.DepartureHarbour,
		detailsRequest.ShipmentDate, detailsRequest.ArrivalDate, detailsRequest.ClearingDate, detailsRequest.ShippingStatus)
	return err
}

func (s *VehicleService) UpdatePurchaseDetails(vehicleID int64, boughtFromName, boughtFromTitle,
	boughtFromContact, boughtFromAddress, boughtFromOtherContacts, purchaseRemarks,
	lcBank, lcNumber *string, lcCostJPY *float64, purchaseDate *time.Time) error {

	query := `
        UPDATE vehicle_purchases 
        SET bought_from_name = $2,
            bought_from_title = $3,
            bought_from_contact = $4,
            bought_from_address = $5,
            bought_from_other_contacts = $6,
            purchase_remarks = $7,
            lc_bank = $8,
            lc_number = $9,
            lc_cost_jpy = $10,
            purchase_date = $11,
            updated_at = CURRENT_TIMESTAMP
        WHERE vehicle_id = $1
    `

	_, err := s.db.Db.Exec(query, vehicleID, boughtFromName, boughtFromTitle,
		boughtFromContact, boughtFromAddress, boughtFromOtherContacts, purchaseRemarks,
		lcBank, lcNumber, lcCostJPY, purchaseDate)
	return err
}

func (s *VehicleService) UpdateFinancialDetails(vehicleID int64, chargesLKR, ttLKR,
	dutyLKR, clearingLKR, otherExpensesLKR *float64, totalCostLKR float64) error {

	query := `
        UPDATE vehicle_financials 
        SET charges_lkr = $2,
            tt_lkr = $3,
            duty_lkr = $4,
            clearing_lkr = $5,
            other_expenses_lkr = $6,
            total_cost_lkr = $7,
            updated_at = CURRENT_TIMESTAMP
        WHERE vehicle_id = $1
    `

	_, err := s.db.Db.Exec(query, vehicleID, chargesLKR, ttLKR, dutyLKR,
		clearingLKR, otherExpensesLKR, totalCostLKR)
	return err
}

func (s *VehicleService) UpdateSalesDetails(vehicleID int64, req request.SalesDetailsRequest) error {

	query := `
        UPDATE vehicle_sales 
        SET sold_date = $2,
            revenue = $3,
            profit = $4,
            sold_to_name = $5,
            sold_to_title = $6,
            contact_number = $7,
            customer_address = $8,
            other_contacts = $9,
            sale_remarks = $10,
            sale_status = $11,
            updated_at = CURRENT_TIMESTAMP
        WHERE vehicle_id = $1
    `

	_, err := s.db.Db.Exec(query, vehicleID, req.SoldDate, req.Revenue, req.Profit, req.SoldToName,
		req.SoldToTitle, req.ContactNumber, req.CustomerAddress, req.OtherContacts, req.SaleRemarks, req.SaleStatus)
	return err
}

func (s *VehicleService) UpdateVehicleDetails(vehicleID int64, req *request.UpdateVehicleRequest) error {

	query := `
        UPDATE vehicles 
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

	_, err := s.db.Db.Exec(query, vehicleID, req.Code, req.Make, req.Model, req.TrimLevel,
		req.YearOfManufacture, req.Color, req.MileageKm, req.ChassisID, req.ConditionStatus,
		req.YearOfRegistration, req.LicensePlate, req.AuctionGrade, req.AuctionPrice, req.CIFValue,
		req.Currency, req.HSCode, req.InvoiceFOBJPY, req.RegistrationNumber, req.RecordDate)
	return err
}

func (s *VehicleService) CreateVehicleMake(vehicleMake request.CreateVehicleMake) (*entity.VehicleMake, error) {
	query := `
        INSERT INTO vehicle_makes (make_name, country_origin, is_active)
        VALUES ($1, $2, $3)
        RETURNING id, make_name, country_origin, is_active, created_at
    `

	var make entity.VehicleMake
	err := s.db.Db.QueryRow(query, vehicleMake.MakeName, vehicleMake.CountryOrigin, vehicleMake.IsActive).Scan(
		&make.ID, &make.MakeName, &make.CountryOrigin, &make.IsActive, &make.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &make, nil
}

func (s *VehicleService) GetAllVehicleMakes(activeOnly bool) ([]entity.VehicleMake, error) {
	query := `SELECT id, make_name, country_origin, is_active, created_at FROM vehicle_makes`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY make_name`

	rows, err := s.db.Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var makes []entity.VehicleMake
	for rows.Next() {
		var make entity.VehicleMake
		err := rows.Scan(&make.ID, &make.MakeName, &make.CountryOrigin, &make.IsActive, &make.CreatedAt)
		if err != nil {
			return nil, err
		}
		makes = append(makes, make)
	}

	return makes, nil
}

func (s *VehicleService) UpdateVehicleMake(id int, makeName, countryOrigin *string, isActive *bool) error {
	query := `
        UPDATE vehicle_makes 
        SET make_name = COALESCE($2, make_name),
            country_origin = COALESCE($3, country_origin),
            is_active = COALESCE($4, is_active)
        WHERE id = $1
    `

	_, err := s.db.Db.Exec(query, id, makeName, countryOrigin, isActive)
	return err
}

// Vehicle Model Service Methods
func (s *VehicleService) CreateVehicleModel(req request.CreateVehicleModel) (*entity.VehicleModel, error) {

	query := `
        INSERT INTO vehicle_models (make_id, model_name, body_type, fuel_type, transmission_type, engine_size_cc, is_active)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, make_id, model_name, body_type, fuel_type, transmission_type, engine_size_cc, is_active, created_at
    `

	var model entity.VehicleModel
	err := s.db.Db.QueryRow(query, req.MakeID, req.ModelName, req.BodyType, req.FuelType, req.TransmissionType, req.EngineSizeCC, req.IsActive).Scan(
		&model.ID, &model.MakeID, &model.ModelName, &model.BodyType, &model.FuelType,
		&model.TransmissionType, &model.EngineSizeCC, &model.IsActive, &model.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (s *VehicleService) GetVehicleModels(makeID *int, activeOnly bool) ([]entity.VehicleModelWithMake, error) {
	query := `
        SELECT vm.id, vm.make_id, vm.model_name, vm.body_type, vm.fuel_type, 
               vm.transmission_type, vm.engine_size_cc, vm.is_active, vm.created_at,
               vma.make_name
        FROM vehicle_models vm
        JOIN vehicle_makes vma ON vm.make_id = vma.id
    `

	var args []interface{}
	var conditions []string

	if makeID != nil {
		conditions = append(conditions, "vm.make_id = $1")
		args = append(args, *makeID)
	}

	if activeOnly {
		if len(args) > 0 {
			conditions = append(conditions, "vm.is_active = $2")
		} else {
			conditions = append(conditions, "vm.is_active = $1")
		}
		args = append(args, true)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY vma.make_name, vm.model_name"

	rows, err := s.db.Db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []entity.VehicleModelWithMake
	for rows.Next() {
		var model entity.VehicleModelWithMake
		err := rows.Scan(&model.ID, &model.MakeID, &model.ModelName, &model.BodyType,
			&model.FuelType, &model.TransmissionType, &model.EngineSizeCC, &model.IsActive,
			&model.CreatedAt, &model.MakeName)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

func (s *VehicleService) GetVehicleModelByID(id int) (*entity.VehicleModelWithMake, error) {
	query := `
        SELECT vm.id, vm.make_id, vm.model_name, vm.body_type, vm.fuel_type, 
               vm.transmission_type, vm.engine_size_cc, vm.is_active, vm.created_at,
               vma.make_name
        FROM vehicle_models vm
        JOIN vehicle_makes vma ON vm.make_id = vma.id
        WHERE vm.id = $1
    `

	var model entity.VehicleModelWithMake
	err := s.db.Db.QueryRow(query, id).Scan(
		&model.ID, &model.MakeID, &model.ModelName, &model.BodyType,
		&model.FuelType, &model.TransmissionType, &model.EngineSizeCC,
		&model.IsActive, &model.CreatedAt, &model.MakeName,
	)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (s *VehicleService) UpdateVehicleModel(id int, modelName, bodyType, fuelType,
	transmissionType *string, engineSizeCC *int, isActive *bool) error {

	query := `
        UPDATE vehicle_models 
        SET model_name = COALESCE($2, model_name),
            body_type = COALESCE($3, body_type),
            fuel_type = COALESCE($4, fuel_type),
            transmission_type = COALESCE($5, transmission_type),
            engine_size_cc = COALESCE($6, engine_size_cc),
            is_active = COALESCE($7, is_active)
        WHERE id = $1
    `

	_, err := s.db.Db.Exec(query, id, modelName, bodyType, fuelType, transmissionType, engineSizeCC, isActive)
	return err
}
