package queryBuilder

import (
	"fmt"
)

type MaxRangeCondition struct {
	field string
	index int
	value interface{}
}

func NewMaxRangeCondition(field string, index int, value interface{}) Condition {
	return &MaxRangeCondition{field, index, value}
}

func (mc *MaxRangeCondition) assemble() string {
	valuePlaceholder := fmt.Sprintf("$%d", mc.index)
	return fmt.Sprintf("%s <= %s", mc.field, valuePlaceholder)
}
