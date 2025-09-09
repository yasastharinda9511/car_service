package filters

import "net/http"

type Filter interface {
	GetValuesFromRequest(*http.Request) Filter
	GetQuery(baseQuery string, orderBy string, limit int, offset int) (string, []interface{})
}
