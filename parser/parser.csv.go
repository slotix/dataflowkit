package parser

import (
	"bytes"
	"encoding/csv"

	"github.com/spf13/viper"
)

//MarshalCSV Marshales harvested data as CSV tables
func (cols Collections) MarshalCSV() ([]byte, error) {

	tables := CSVTableCollection{}
	for _, o := range cols.Element {
		tables.Tables = append(tables.Tables, o.marshalCSVItem())
	}
	buf, err := tables.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (c collection) marshalCSVItem() CSVTable {
	var b bytes.Buffer
	writer := csv.NewWriter(&b)
	str := viper.GetString("parser.CSV.comma")
	r := rune(str[0])
	writer.Comma = r

	buf := c.generateTable()

	writer.WriteAll(buf)
	writer.Flush()
	var table CSVTable
	table.URL = c.URL
	table.Content = string(b.Bytes())
	return table
}
