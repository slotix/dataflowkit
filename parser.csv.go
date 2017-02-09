package parser

import (
	"bytes"
	"encoding/csv"
)

//MarshalCSV Marshales harvested data as CSV tables
func (out Out) MarshalCSV() ([]byte, error) {

	tables := CSVTableCollection{}
	for _, o := range out.Element {
		tables.Tables = append(tables.Tables, o.marshalCSVItem())
	}
	buf, err := tables.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (o outItem) marshalCSVItem() CSVTable {
	var b bytes.Buffer
	writer := csv.NewWriter(&b)
	//str := viper.GetString("parser.CSV.comma")
	//r := rune(str[0])
	//writer.Comma = r

	buf := o.generateTable()

	writer.WriteAll(buf)
	writer.Flush()
	var table CSVTable
	table.URL = o.URL
	table.Content = string(b.Bytes())
	return table
}
