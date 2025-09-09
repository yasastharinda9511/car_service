package queryBuilder

import (
	"fmt"
)

type LikeCondition struct {
	field string
	index int
	value interface{}
}

func NewLikeCondition(condition string, index int, value interface{}) Condition {
	return &LikeCondition{condition, index, value}
}

func (lc *LikeCondition) assemble() string {
	placeholder := fmt.Sprintf("$%d", lc.index)
	condition := fmt.Sprintf("%s ILIKE %s", lc.field, placeholder)
	return condition
}
