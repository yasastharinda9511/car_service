package queryBuilder

import (
	"fmt"
)

type MinRangeCondition struct {
	field string
	index int
	value interface{}
}

func NewMinRangeCondition(field string, index int, value interface{}) Condition {
	return &MinRangeCondition{field, index, value}
}

func (mc *MinRangeCondition) assemble() string {
	valuePlaceholder := fmt.Sprintf("$%d", mc.index)
	return fmt.Sprintf("%s >= %s", mc.field, valuePlaceholder)
}
