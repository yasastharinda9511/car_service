package queryBuilder

import (
	"fmt"
	"strings"
)

// QueryBuilder helps build dynamic SQL queries
type QueryBuilder struct {
	baseQuery  string
	conditions []Condition
	args       []interface{}
	orderBy    *OrderBy
	argCounter int
}

// NewQueryBuilder creates a new query builder with base query
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		conditions: make([]Condition, 0),
		args:       make([]interface{}, 0),
		orderBy:    nil,
		argCounter: 0,
	}
}

// AddCondition adds a WHERE condition
func (qb *QueryBuilder) AddCondition(condition string, arg interface{}) {
	qb.argCounter++
	// Replace placeholder with actual parameter number
	generalCondition := NewGeneralCondition(condition, qb.argCounter, arg)
	qb.conditions = append(qb.conditions, generalCondition)
	qb.args = append(qb.args, arg)
}

// AddRangeCondition adds a range condition (BETWEEN)
func (qb *QueryBuilder) AddRangeCondition(field string, min interface{}, max interface{}) {
	qb.argCounter++
	minIndex := qb.argCounter
	qb.argCounter++
	maxIndex := qb.argCounter

	condition := NewRangeCondition(field, minIndex, maxIndex, min, max)
	qb.conditions = append(qb.conditions, condition)
	qb.args = append(qb.args, min, max)
}

// AddLikeCondition adds a LIKE condition for text search
func (qb *QueryBuilder) AddLikeCondition(field string, value string) {
	qb.argCounter++
	likeCondition := NewLikeCondition(field, qb.argCounter, value)
	qb.conditions = append(qb.conditions, likeCondition)
	qb.args = append(qb.args, "%"+value+"%")
}

func (qb *QueryBuilder) AddMinRangeCondition(field string, arg interface{}) {
	qb.argCounter++
	// Replace placeholder with actual parameter number
	generalCondition := NewMinRangeCondition(field, qb.argCounter, arg)
	qb.conditions = append(qb.conditions, generalCondition)
	qb.args = append(qb.args, arg)
}

func (qb *QueryBuilder) AddMaxRangeCondition(field string, arg interface{}) {
	qb.argCounter++
	// Replace placeholder with actual parameter number
	generalCondition := NewMaxRangeCondition(field, qb.argCounter, arg)
	qb.conditions = append(qb.conditions, generalCondition)
	qb.args = append(qb.args, arg)
}

func (qb *QueryBuilder) AddOrderBy(field string, sortingOrder string) {
	qb.orderBy = NewOrderBy(field, sortingOrder)
}

//
//// AddInCondition adds an IN condition for array values
//func (qb *QueryBuilder) AddInCondition(field string, values []string) {
//	if len(values) == 0 {
//		return
//	}
//
//	placeholders := make([]string, len(values))
//	for i, value := range values {
//		qb.argCounter++
//		placeholders[i] = fmt.Sprintf("$%d", qb.argCounter)
//		qb.args = append(qb.args, value)
//	}
//
//	condition := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
//	qb.conditions = append(qb.conditions, condition)
//}

// Build constructs the final query
func (qb *QueryBuilder) Build(baseQuery string, groupBy string, orderBy string, limit int, offset int, skipOrderBy bool) (string, []interface{}) {
	query := baseQuery

	// Add WHERE conditions
	if len(qb.conditions) > 0 {
		placeholders := make([]string, len(qb.conditions))
		for i, condition := range qb.conditions {
			placeholders[i] = condition.assemble()
		}
		query += " WHERE " + strings.Join(placeholders, " AND ")
	}

	// Add GROUP BY
	if groupBy != "" {
		query += " GROUP BY " + groupBy
	}

	// Add ORDER BY (skip for count queries)
	if !skipOrderBy {
		if qb.orderBy != nil {
			query += " " + qb.orderBy.assemble()
		} else if orderBy != "" {
			query += " ORDER BY " + orderBy
		}
	}

	// Add LIMIT and OFFSET
	newargs := append([]interface{}{}, qb.args...)
	if limit > 0 {
		qb.argCounter++
		limitPlaceholder := fmt.Sprintf("$%d", qb.argCounter)
		query += fmt.Sprintf(" LIMIT %s", limitPlaceholder)
		newargs = append(newargs, limit)

	}

	if offset > 0 {
		qb.argCounter++
		offsetPlaceholder := fmt.Sprintf("$%d", qb.argCounter)
		query += fmt.Sprintf(" OFFSET %s", offsetPlaceholder)
		newargs = append(newargs, offset)
	}

	return query, newargs
}
