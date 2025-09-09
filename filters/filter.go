package filters

import "net/http"

type Filter interface {
	GetValuesFromRequest(*http.Request) Filter
	GetQuery(baseQuery string, orderBy string, limit, offset int) (string, []interface{})
}
