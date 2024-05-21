package data

import (
	"fmt"
	"sort"

	"github.com/Malpizarr/dbproto/pkg/dbdata"
	"google.golang.org/protobuf/types/known/structpb"
)

// The Query functionality allows you to perform complex queries on your database table.
// A query can include filters, sorting, limits, and offsets, which help in retrieving specific subsets of data efficiently.
type Query struct {
	Filters map[string]interface{} // Filters to select specific records
	SortBy  string                 // SortBy is a Field to sort the records by
	Limit   int                    // Limit is the Maximum number of records to return
	Offset  int                    // Offset is the Number of records to skip (for pagination)
}

// ExecutionPlan represents the execution plan of a query.
type ExecutionPlan struct {
	IndexToUse string
	Filters    map[string]interface{}
	SortBy     string
	Limit      int
	Offset     int
}

// selectBestIndex selects the best index for a given query.
func (t *Table) selectBestIndex(query Query) string {
	bestIndex := ""
	bestSelectivity := 1.0 // Worst possible selectivity

	// Iterate over each filter field to find the best index
	for field := range query.Filters {
		if index, exists := t.Indexes[field]; exists {
			selectivity := float64(len(index)) / float64(len(t.Records))
			if selectivity < bestSelectivity {
				bestSelectivity = selectivity
				bestIndex = field
			}
		}
	}

	return bestIndex
}

// generateExecutionPlan generates an execution plan for a given query.
func (t *Table) generateExecutionPlan(query Query) ExecutionPlan {
	bestIndex := t.selectBestIndex(query)
	return ExecutionPlan{
		IndexToUse: bestIndex,
		Filters:    query.Filters,
		SortBy:     query.SortBy,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}
}

// executePlan executes the execution plan and returns the resulting records.
func (t *Table) executePlan(plan ExecutionPlan) ([]*dbdata.Record, error) {
	var results []*dbdata.Record

	// If an index is used, search within the indexed records
	if plan.IndexToUse != "" {
		for _, record := range t.Indexes[plan.IndexToUse] {
			if match(record, plan.Filters) {
				results = append(results, record)
			}
		}
	} else {
		// Otherwise, search within all records
		for _, record := range t.Records {
			if match(record, plan.Filters) {
				results = append(results, record)
			}
		}
	}

	// Sort the results if a sort field is specified
	if plan.SortBy != "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Fields[plan.SortBy].GetNumberValue() < results[j].Fields[plan.SortBy].GetNumberValue()
		})
	} else {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Fields[t.PrimaryKey].GetStringValue() < results[j].Fields[t.PrimaryKey].GetStringValue()
		})
	}

	// Apply offset to the results
	if plan.Offset > 0 {
		if plan.Offset >= len(results) {
			return []*dbdata.Record{}, nil
		}
		results = results[plan.Offset:]
	}

	// Apply limit to the results
	if plan.Limit > 0 && plan.Limit < len(results) {
		results = results[:plan.Limit]
	}

	return results, nil
}

// match checks if a record matches the given filters.
func match(record *dbdata.Record, filters map[string]interface{}) bool {
	for field, value := range filters {
		protoValue, err := structpb.NewValue(value)
		if err != nil {
			fmt.Printf("Error converting filter value for field %s: %v\n", field, err)
			return false
		}
		recordValue, exists := record.Fields[field]
		if !exists {
			fmt.Printf("Field %s does not exist in record\n", field)
			return false
		}
		if !Equal(recordValue, protoValue) {
			return false
		}
	}
	return true
}

// Query is a method of the Table struct that performs a query on the table and returns the resulting records.
// It first generates an execution plan for the given query.
// The execution plan includes the best index to use for the query, the filters to apply, the field to sort by, and the limit and offset for the results.
// It then executes the execution plan, searching for records that match the filters, sorting the results, and applying the limit and offset.
// If an error occurs during the execution of the plan, it returns the error.
//
// Parameters:
// - query: A Query object representing the query to be performed. The Query object includes filters to select specific records, a field to sort the records by, and a limit and offset for the results.
//
// Returns:
// - A slice of pointers to dbdata.Record objects, representing the records that match the query. If no records match the query, it returns an empty slice.
// - An error, if any error occurs during the query operation. If the operation is successful, the error is nil.
func (t *Table) Query(query Query) ([]*dbdata.Record, error) {
	plan := t.generateExecutionPlan(query)
	return t.executePlan(plan)
}
