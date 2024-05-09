package exports

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"

	"github.com/Malpizarr/dbproto/pkg/dbdata"
	"google.golang.org/protobuf/types/known/structpb"
)

func formatProtoValueCSV(val *structpb.Value) string {
	if val == nil {
		return ""
	}
	switch x := val.Kind.(type) {
	case *structpb.Value_StringValue:
		return x.StringValue
	case *structpb.Value_NumberValue:
		return fmt.Sprintf("%g", x.NumberValue)
	case *structpb.Value_BoolValue:
		return fmt.Sprintf("%t", x.BoolValue)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func ExportRecordsToCSV(records []*dbdata.Record, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	keySet := make(map[string]bool)
	for _, rec := range records {
		for key := range rec.Fields {
			keySet[key] = true
		}
	}

	headers := make([]string, 0, len(keySet))
	for key := range keySet {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, rec := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			if val, ok := rec.Fields[header]; ok && val != nil {
				row[i] = formatProtoValueCSV(val)
			} else {
				row[i] = ""
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
