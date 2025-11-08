package filters

import (
	"car_service/queryBuilder"
	"net/http"
	"time"
)

type VehicleSalesFilter struct {
	DateFrom     *time.Time
	DateTo       *time.Time
	QueryBuilder *queryBuilder.QueryBuilder
}

func NewVehicleSalesFilter() Filter {
	return &VehicleSalesFilter{QueryBuilder: queryBuilder.NewQueryBuilder()}
}
func (v *VehicleSalesFilter) GetValuesFromRequest(r *http.Request) Filter {

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
		v.QueryBuilder.AddRangeCondition("vsl.created_at", *v.DateFrom, *v.DateTo)
	} else if v.DateTo != nil {
		v.QueryBuilder.AddMinRangeCondition("vsl.created_at", *v.DateTo)
	} else if v.DateFrom != nil {
		v.QueryBuilder.AddMaxRangeCondition("vsl.created_at", *v.DateFrom)
	}

	return v
}

func (v *VehicleSalesFilter) GetQuery(baseQuery string, groupBy string, orderBy string, limit, offset int) (string, []interface{}) {
	return v.QueryBuilder.Build(baseQuery, groupBy, orderBy, limit, offset, false)
}

func (v *VehicleSalesFilter) GetQueryForCount(baseQuery string, groupBy string, orderBy string, limit int, offset int) (string, []interface{}) {
	return v.QueryBuilder.Build(baseQuery, groupBy, orderBy, limit, offset, true)
}
