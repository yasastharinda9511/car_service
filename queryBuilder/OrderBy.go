package queryBuilder

import (
	"fmt"
	"strings"
)

type OrderBy struct {
	field        string
	sortingOrder string
}

func NewOrderBy(field string, sortingOrder string) *OrderBy {
	return &OrderBy{field, sortingOrder}
}

func (ob *OrderBy) assemble() string {
	// Normalize sorting order
	order := strings.ToUpper(strings.TrimSpace(ob.sortingOrder))
	if order != "ASC" && order != "DESC" {
		order = "ASC" // default fallback
	}

	return fmt.Sprintf("ORDER BY %s %s", ob.field, order)
}
