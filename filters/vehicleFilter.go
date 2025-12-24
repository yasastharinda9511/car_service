package filters

import (
	"car_service/queryBuilder"
	"net/http"
	"strconv"
	"time"
)

type VehicleFilters struct {
	Make            string
	Model           string
	Year            int
	YearMin         int
	YearMax         int
	Color           string
	ConditionStatus string
	ShippingStatus  string
	SaleStatus      string
	MileageMin      int
	MileageMax      int
	PriceMin        float64
	PriceMax        float64
	Makes           []string
	Models          []string
	Colors          []string
	Search          string
	DateFrom        *time.Time
	DateTo          *time.Time
	QueryBuilder    *queryBuilder.QueryBuilder
}

func NewVehicleFilters() Filter {
	return &VehicleFilters{QueryBuilder: queryBuilder.NewQueryBuilder()}
}

func (v *VehicleFilters) GetValuesFromRequest(r *http.Request) Filter {

	v.Make = r.URL.Query().Get("make")
	if v.Make != "" {
		v.QueryBuilder.AddCondition(GetMappedField("make"), v.Make)
	}

	v.Model = r.URL.Query().Get("model")
	if v.Model != "" {
		v.QueryBuilder.AddCondition(GetMappedField("model"), v.Model)
	}

	v.ConditionStatus = r.URL.Query().Get("condition_status")
	if v.ConditionStatus != "" {
		v.QueryBuilder.AddCondition(GetMappedField("condition_status"), v.ConditionStatus)
	}

	v.ShippingStatus = r.URL.Query().Get("shipping_status")
	if v.ShippingStatus != "" {
		v.QueryBuilder.AddCondition(GetMappedField("shipping_status"), v.ShippingStatus)
	}

	v.SaleStatus = r.URL.Query().Get("sale_status")
	if v.SaleStatus != "" {
		v.QueryBuilder.AddCondition(GetMappedField("sale_status"), v.SaleStatus)
	}

	if year := r.URL.Query().Get("year"); year != "" {
		v.Year, _ = strconv.Atoi(year)
		v.QueryBuilder.AddCondition(GetMappedField("year"), v.Year)
	}

	v.Search = r.URL.Query().Get("search")
	if v.Search != "" {
		v.QueryBuilder.AddLikeCondition("(v.make || ' ' || v.model || ' ' || v.chassis_id || ' ' || v.color || ' ' || vs.shipping_status || ' ' || vsl.sale_status || ' ' || CAST(v.code AS TEXT) || ' ' || COALESCE(v.trim_level, '') || ' ' || CAST(v.year_of_manufacture AS TEXT) || ' ' || COALESCE(v.license_plate, '') || ' ' || COALESCE(v.auction_grade, '') || ' ' || COALESCE(vs.vessel_name, '') || ' ' || COALESCE(vs.departure_harbour, '') || ' ' || COALESCE(c.customer_name, '') || ' ' || COALESCE(vp.bought_from_name, ''))", v.Search)
	}

	if mileageMin := r.URL.Query().Get("mileage_min"); mileageMin != "" {
		v.MileageMin, _ = strconv.Atoi(mileageMin)
	}

	if mileageMax := r.URL.Query().Get("mileage_max"); mileageMax != "" {
		v.MileageMax, _ = strconv.Atoi(mileageMax)
	}

	if v.MileageMin != 0 && v.MileageMax != 0 {
		v.QueryBuilder.AddRangeCondition(GetMappedField("mileage"), v.MileageMin, v.MileageMax)
	} else if v.MileageMin != 0 {
		v.QueryBuilder.AddMinRangeCondition(GetMappedField("mileage"), v.MileageMin)
	} else if v.MileageMax != 0 {
		v.QueryBuilder.AddMaxRangeCondition(GetMappedField("mileage"), v.MileageMax)
	}

	if yearMin := r.URL.Query().Get("year_min"); yearMin != "" {
		v.YearMin, _ = strconv.Atoi(yearMin)
	}
	if yearMax := r.URL.Query().Get("year_max"); yearMax != "" {
		v.YearMax, _ = strconv.Atoi(yearMax)
	}
	if (v.YearMin != 0 && v.YearMax != 0) && v.Year == 0 {
		v.QueryBuilder.AddRangeCondition(GetMappedField("year"), v.YearMin, v.YearMax)
	} else if v.YearMin != 0 && v.YearMax == 0 {
		v.QueryBuilder.AddMinRangeCondition(GetMappedField("year"), v.YearMin)
	} else if v.YearMax != 0 && v.Year == 0 {
		v.QueryBuilder.AddMaxRangeCondition(GetMappedField("year"), v.YearMax)
	}

	v.Color = r.URL.Query().Get("color")
	if v.Color != "" {
		v.QueryBuilder.AddCondition(GetMappedField("color"), v.Color)
	}

	dateFromStr := r.URL.Query().Get("dateRangeStart")
	if dateFromStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateFromStr)
		if err == nil {
			v.DateFrom = &parsedDate
		}

	}

	dateToStr := r.URL.Query().Get("dateRangeEnd")
	if dateToStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateToStr)
		if err == nil {
			v.DateTo = &parsedDate
		}
	}

	if v.DateFrom != nil && v.DateTo != nil {
		v.QueryBuilder.AddRangeCondition(GetMappedField("created_at"), *v.DateFrom, *v.DateTo)
	} else if v.DateTo != nil {
		v.QueryBuilder.AddMinRangeCondition(GetMappedField("created_at"), *v.DateTo)
	} else if v.DateFrom != nil {
		v.QueryBuilder.AddMaxRangeCondition(GetMappedField("created_at"), *v.DateFrom)
	}

	orderBy := r.URL.Query().Get("order_by")
	sort := r.URL.Query().Get("sort")
	if orderBy != "" {
		// Validate and map the orderBy field
		if IsValidOrderByField(orderBy) {
			if sort == "" {
				sort = "ASC" // default sort order
			}
			// Map the user-friendly field name to database field with alias
			mappedField := GetMappedField(orderBy)
			v.QueryBuilder.AddOrderBy(mappedField, sort)
		}
		// If invalid field, silently ignore (or you can return an error)
	}

	return v
}

func (v *VehicleFilters) GetQuery(baseQuery string, groupBy string, orderBy string, limit, offset int) (string, []interface{}) {
	return v.QueryBuilder.Build(baseQuery, groupBy, orderBy, limit, offset, false)
}

func (v *VehicleFilters) GetQueryForCount(baseQuery string, groupBy string, orderBy string, limit, offset int) (string, []interface{}) {
	return v.QueryBuilder.Build(baseQuery, groupBy, orderBy, limit, offset, true)
}
