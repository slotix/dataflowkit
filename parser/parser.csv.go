package parser

import (
	"bytes"
	"encoding/csv"
)

//MarshalCSV Marshales harvested data as CSV tables
func (cols Collections) MarshalCSV(comma string) ([]byte, error) {

	tables := CSVTableCollection{}
	for _, o := range cols.Collections {
		tables.Tables = append(tables.Tables, o.marshalCSVItem(comma))
	}
	buf, err := tables.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (c collection) marshalCSVItem(comma string) CSVTable {
	if comma == "" {
		comma = ","
	}
	var b bytes.Buffer
	writer := csv.NewWriter(&b)
	r := rune(comma[0])
	writer.Comma = r

	buf := c.generateTable()

	writer.WriteAll(buf)
	writer.Flush()
	var table CSVTable
	table.URL = c.URL
	table.Content = string(b.Bytes())
	return table
}
