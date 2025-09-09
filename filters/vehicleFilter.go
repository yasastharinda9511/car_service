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
	DateFrom        time.Time
	DateTo          time.Time
	QueryBuilder    *queryBuilder.QueryBuilder
}

func NewVehicleFilters() Filter {
	return &VehicleFilters{QueryBuilder: queryBuilder.NewQueryBuilder()}
}

func (v *VehicleFilters) GetValuesFromRequest(r *http.Request) Filter {

	v.Make = r.URL.Query().Get("make")
	if v.Make != "" {
		v.QueryBuilder.AddCondition("v.make", v.Make)
	}

	v.Model = r.URL.Query().Get("model")
	if v.Model != "" {
		v.QueryBuilder.AddCondition("v.model", v.Model)
	}

	v.ConditionStatus = r.URL.Query().Get("condition_status")
	if v.ConditionStatus != "" {
		v.QueryBuilder.AddCondition("v.condition_status", v.ConditionStatus)
	}

	v.ShippingStatus = r.URL.Query().Get("shipping_status")
	if v.ConditionStatus != "" {
		v.QueryBuilder.AddCondition("vp.shipping_status", v.ShippingStatus)
	}

	v.SaleStatus = r.URL.Query().Get("sale_status")
	if v.SaleStatus != "" {
		v.QueryBuilder.AddCondition("v.sale_status", v.SaleStatus)
	}

	if year := r.URL.Query().Get("year"); year != "" {
		v.Year, _ = strconv.Atoi(year)
		v.QueryBuilder.AddCondition("v.year_of_manufacture", v.Year)
	}

	v.Search = r.URL.Query().Get("search")
	if v.Search != "" {
		v.QueryBuilder.AddLikeCondition("(v.make || ' ' || v.model || ' ' || v.chassis_id || ' ' || v.color || ' ' || vs.shipping_status || ' ' || vsl.sale_status)", v.Search)
	}

	if mileageMin := r.URL.Query().Get("mileage_min"); mileageMin != "" {
		v.MileageMin, _ = strconv.Atoi(mileageMin)
	}

	if mileageMax := r.URL.Query().Get("mileage_max"); mileageMax != "" {
		v.MileageMax, _ = strconv.Atoi(mileageMax)
	}

	if v.MileageMin != 0 && v.MileageMax != 0 {
		v.QueryBuilder.AddRangeCondition("v.mileage_km", v.MileageMin, v.MileageMax)
	} else if v.MileageMin != 0 {
		v.QueryBuilder.AddMinRangeCondition("v.mileage_km", v.MileageMin)
	} else if v.MileageMax != 0 {
		v.QueryBuilder.AddMaxRangeCondition("v.mileage_km", v.MileageMax)
	}

	if yearMin := r.URL.Query().Get("year_min"); yearMin != "" {
		v.YearMin, _ = strconv.Atoi(yearMin)
	}
	if yearMax := r.URL.Query().Get("year_max"); yearMax != "" {
		v.YearMax, _ = strconv.Atoi(yearMax)
	}
	if (v.YearMin != 0 && v.YearMax != 0) && v.Year == 0 {
		v.QueryBuilder.AddRangeCondition("v.year_of_manufacture", v.MileageMin, v.MileageMax)
	} else if v.YearMin != 0 && v.YearMax == 0 {
		v.QueryBuilder.AddMinRangeCondition("v.year_of_manufacture", v.YearMin)
	} else if v.YearMax != 0 && v.Year == 0 {
		v.QueryBuilder.AddMaxRangeCondition("v.year_of_manufacture", v.YearMax)
	}

	v.Color = r.URL.Query().Get("color")
	if v.Color != "" {
		v.QueryBuilder.AddCondition("v.color", v.Color)
	}

	return v
}

func (v *VehicleFilters) GetQuery(baseQuery string, orderBy string, limit, offset int) (string, []interface{}) {
	return v.QueryBuilder.Build(baseQuery, orderBy, limit, offset)
}
