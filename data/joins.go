package data

import (
	"fmt"

	"github.com/Malpizarr/dbproto/dbdata"
)

type JoinType int

const (
	InnerJoin JoinType = iota
	LeftJoin
	RightJoin
	FullOuterJoin
)

func JoinTables(t1, t2 *Table, key1, key2 string, joinType JoinType) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0)

	if err := t1.ResetAndLoadIndexes(); err != nil {
		return nil, fmt.Errorf("failed to load indexes for table 1: %v", err)
	}
	if err := t2.ResetAndLoadIndexes(); err != nil {
		return nil, fmt.Errorf("failed to load indexes for table 2: %v", err)
	}

	for _, rec1 := range t1.Indexes[key1] {
		if rec1 == nil {
			continue
		}

		for _, rec2 := range t2.Indexes[key2] {
			if rec2 != nil {
				results = append(results, mergeRecords(rec1, rec2))
			}
		}

		if len(t2.Indexes[key2]) == 0 && (joinType == LeftJoin || joinType == FullOuterJoin) {
			results = append(results, mergeRecords(rec1, nil))
		}
	}

	if joinType == RightJoin || joinType == FullOuterJoin {
		for _, rec2 := range t2.Indexes[key2] {
			if rec2 == nil {
				continue
			}

			if len(t1.Indexes[key1]) == 0 {
				results = append(results, mergeRecords(nil, rec2))
			}
		}
	}

	return results, nil
}

func mergeRecords(rec1, rec2 *dbdata.Record) map[string]interface{} {
	result := make(map[string]interface{})
	if rec1 != nil {
		for k, v := range rec1.Fields {
			result["t1."+k] = v
		}
	}
	if rec2 != nil {
		for k, v := range rec2.Fields {
			result["t2."+k] = v
		}
	}
	return result
}
