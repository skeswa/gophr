package query

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/gophr-pm/gophr/lib/db"
)

// SelectQueryBuilder constructs a select query.
type SelectQueryBuilder struct {
	table          string
	limit          *int
	columns        []string
	conditions     []*Condition
	allowFiltering bool
}

// Select starts constructing a select query.
func Select(columns ...string) *SelectQueryBuilder {
	return &SelectQueryBuilder{
		columns: columns,
	}
}

// SelectSum starts constructing a select query that sums a specific column of
// the rows in the result set.
func SelectSum(column string) *SelectQueryBuilder {
	return &SelectQueryBuilder{
		columns: []string{fmt.Sprintf(sumOperatorTemplate, column)},
	}
}

// SelectCount starts constructing a select query that counts rows.
func SelectCount() *SelectQueryBuilder {
	return &SelectQueryBuilder{columns: []string{countOperator}}
}

// From specifies the table in a select query.
func (qb *SelectQueryBuilder) From(table string) *SelectQueryBuilder {
	qb.table = table
	return qb
}

// Where adds a condition to which all of the selected rows should adhere.
func (qb *SelectQueryBuilder) Where(condition *Condition) *SelectQueryBuilder {
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// And is an alias for SelectQueryBuilder.Where(condition).
func (qb *SelectQueryBuilder) And(condition *Condition) *SelectQueryBuilder {
	return qb.Where(condition)
}

// Limit specifies the maximum number of results to fetch.
func (qb *SelectQueryBuilder) Limit(limit int) *SelectQueryBuilder {
	limitClone := limit
	qb.limit = &limitClone
	return qb
}

// AllowFiltering signals that filtering is allowed on the query results.
func (qb *SelectQueryBuilder) AllowFiltering() *SelectQueryBuilder {
	qb.allowFiltering = true
	return qb
}

// Create serializes and creates the query.
func (qb *SelectQueryBuilder) Create(q db.Queryable) db.Query {
	var (
		buffer     bytes.Buffer
		parameters []interface{}
	)

	buffer.WriteString("select ")
	for i, column := range qb.columns {
		if i > 0 {
			buffer.WriteByte(',')
		}

		buffer.WriteString(column)
	}
	buffer.WriteString(" from ")
	buffer.WriteString(DBKeyspaceName)
	buffer.WriteByte('.')
	buffer.WriteString(qb.table)
	if qb.conditions != nil {
		buffer.WriteString(" where ")
		for i, cond := range qb.conditions {
			if i > 0 {
				buffer.WriteString(" and ")
			}

			if cond.hasParameter {
				parameters = append(parameters, cond.parameter)
			}

			buffer.WriteString(cond.expression)
		}
	}
	if qb.limit != nil {
		buffer.WriteString(" limit ")
		buffer.WriteString(strconv.Itoa(*qb.limit))
	}
	if qb.allowFiltering {
		buffer.WriteString(" allow filtering")
	}

	return q.Query(buffer.String(), parameters...)
}
