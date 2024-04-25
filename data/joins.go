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
	var results []map[string]interface{}

	if err := t1.LoadIndexes(); err != nil {
		return nil, err
	}
	if err := t2.LoadIndexes(); err != nil {
		return nil, err
	}

	index1, ok1 := t1.Indexes[key1]
	index2, ok2 := t2.Indexes[key2]

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("one of the keys %s or %s does not have an index", key1, key2)
	}

	keysInLeft := make(map[string]bool)
	keysInRight := make(map[string]bool)

	for key, rec1 := range index1 {
		keysInLeft[key] = true
		if rec2, exists := index2[key]; exists {
			result := mergeRecords(rec1, rec2)
			results = append(results, result)
			keysInRight[key] = true
		} else if joinType == LeftJoin || joinType == FullOuterJoin {
			result := mergeRecords(rec1, nil)
			results = append(results, result)
		}
	}

	if joinType == RightJoin || joinType == FullOuterJoin {
		for key, rec2 := range index2 {
			if !keysInLeft[key] {
				result := mergeRecords(nil, rec2)
				results = append(results, result)
				keysInRight[key] = true
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
