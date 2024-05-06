package data

import (
	"fmt"

	"github.com/Malpizarr/dbproto/pkg/dbdata"
	"google.golang.org/protobuf/types/known/structpb"
)

type JoinType int

const (
	InnerJoin JoinType = iota
	LeftJoin
	RightJoin
	FullOuterJoin
)

// Func to create joins  between two tables it perform one to many join between two tables, based on the key fields provided

func JoinTables(t1, t2 *Table, key1, key2 string, joinType JoinType) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0)

	if err := t1.ResetAndLoadIndexes(); err != nil {
		return nil, fmt.Errorf("failed to load indexes for table 1: %v", err)
	}
	if err := t2.ResetAndLoadIndexes(); err != nil {
		return nil, fmt.Errorf("failed to load indexes for table 2: %v", err)
	}

	// Process records from t1
	for _, rec1 := range t1.Indexes[key1] {
		if rec1 == nil {
			continue
		}

		// Attempt to find matching records in t2
		matched := false
		for _, rec2 := range t2.Indexes[key2] {
			if rec2 != nil && isEqual(rec1.Fields[key1], rec2.Fields[key2]) {
				results = append(results, mergeRecords(rec1, rec2))
				matched = true
			}
		}

		// If no match found and it's a left join or full outer join, add rec1 alone
		if !matched && (joinType == LeftJoin || joinType == FullOuterJoin) {
			results = append(results, mergeRecords(rec1, nil))
		}
	}

	// Process records from t2 if it's a right join or full outer join
	if joinType == RightJoin || joinType == FullOuterJoin {
		for _, rec2 := range t2.Indexes[key2] {
			if rec2 == nil {
				continue
			}

			// Check if rec2 was matched
			matched := false
			for _, rec1 := range t1.Indexes[key1] {
				if rec1 != nil && isEqual(rec1.Fields[key1], rec2.Fields[key2]) {
					matched = true
					break
				}
			}

			// If no corresponding rec1 was found, add rec2 alone
			if !matched {
				results = append(results, mergeRecords(nil, rec2))
			}
		}
	}

	return results, nil
}

// Helper function to compare *structpb.Value fields
func isEqual(val1, val2 *structpb.Value) bool {
	if val1 == nil || val2 == nil {
		return false
	}
	// Compare based on actual type; here assuming both are string values for simplicity
	return val1.GetStringValue() == val2.GetStringValue()
}

func mergeRecords(rec1, rec2 *dbdata.Record) map[string]interface{} {
	result := make(map[string]interface{})
	if rec1 != nil {
		for k, v := range rec1.Fields {
			result["t1."+k] = v.GetStringValue()
		}
	}
	if rec2 != nil {
		for k, v := range rec2.Fields {
			result["t2."+k] = v.GetStringValue()
		}
	}
	return result
}
