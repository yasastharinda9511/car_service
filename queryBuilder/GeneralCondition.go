package queryBuilder

import (
	"fmt"
	"strconv"
	"strings"
)

type GeneralCondition struct {
	field string
	index int
	value interface{}
}

func NewGeneralCondition(condition string, index int, value interface{}) Condition {
	return &GeneralCondition{condition, index, value}
}

func (qc *GeneralCondition) assemble() string {
	condition := strings.ReplaceAll(qc.field, "?", fmt.Sprintf("$%d", qc.index))
	return condition + " = " + "$" + strconv.Itoa(qc.index)
}
