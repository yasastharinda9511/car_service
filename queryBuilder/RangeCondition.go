package queryBuilder

import (
	"fmt"
)

type RangeCondition struct {
	field    string
	indexMin int
	indexMax int
	valueMin interface{}
	valueMax interface{}
}

func NewRangeCondition(field string, indexMin int, indexMax int, valueMin interface{}, valueMax interface{}) Condition {
	return &RangeCondition{field, indexMin, indexMax, valueMin, valueMax}
}

func (rc *RangeCondition) assemble() string {
	minPlaceholder := fmt.Sprintf("$%d", rc.indexMin)
	maxPlaceholder := fmt.Sprintf("$%d", rc.indexMax)

	return fmt.Sprintf("%s BETWEEN %s AND %s", rc.field, minPlaceholder, maxPlaceholder)
}
