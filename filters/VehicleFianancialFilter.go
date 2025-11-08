package filters

import (
	"car_service/queryBuilder"
	"net/http"
	"time"
)

type VehicleFinancialFilter struct {
	DateFrom     *time.Time
	DateTo       *time.Time
	QueryBuilder *queryBuilder.QueryBuilder
}

func NewVehicleFinancialFilter() Filter {
	return &VehicleFinancialFilter{QueryBuilder: queryBuilder.NewQueryBuilder()}
}
func (v *VehicleFinancialFilter) GetValuesFromRequest(r *http.Request) Filter {

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
		v.QueryBuilder.AddRangeCondition("vf.created_at", *v.DateFrom, *v.DateTo)
	} else if v.DateTo != nil {
		v.QueryBuilder.AddMinRangeCondition("vf.created_at", *v.DateTo)
	} else if v.DateFrom != nil {
		v.QueryBuilder.AddMaxRangeCondition("vf.created_at", *v.DateFrom)
	}

	return v
}

func (v *VehicleFinancialFilter) GetQuery(baseQuery string, groupBy string, orderBy string, limit, offset int) (string, []interface{}) {
	return v.QueryBuilder.Build(baseQuery, groupBy, orderBy, limit, offset, false)
}

func (v *VehicleFinancialFilter) GetQueryForCount(baseQuery string, groupBy string, orderBy string, limit int, offset int) (string, []interface{}) {
	return v.QueryBuilder.Build(baseQuery, groupBy, orderBy, limit, offset, true)
}
