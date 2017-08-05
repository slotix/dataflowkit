package parse

import (
	"bytes"
	"encoding/csv"
	"strings"

	"github.com/clbanning/mxj"
)

//encodeCSV writes data to w *csv.Writer.
//header - headers for csv.
//includeHeader include headers or not.
//rows - csv records to be written.
func encodeCSV(header []string, includeHeader bool, rows []map[string]interface{}, comma string, w *csv.Writer) error {
	if comma == "" {
		comma = ","
	}
	w.Comma = rune(comma[0])
	//Add Header string to csv or no
	if includeHeader {
		if err := w.Write(header); err != nil {
			return err
		}
	}
	r := make([]string, len(header))
	for _, row := range rows {
		for i, column := range header {
			switch v := row[column].(type) {
			case string:
				r[i] = v
			case []string:
				r[i] = strings.Join(v, ";")
			case nil:
				r[i] = ""
			}
		}
		if err := w.Write(r); err != nil {
			return err
		}
	}
	return nil
}

//encodeXML writes data blocks to XML.
func encodeXML(blocks []map[string]interface{}, buf *bytes.Buffer) error {
	mxj.XMLEscapeChars(true)
	//write header to xml
	buf.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
	buf.Write([]byte("<items>"))
	for _, piece := range blocks {
		m := mxj.Map(piece)
		//err := m.XmlIndentWriter(&buf, "", "  ", "object")
		err := m.XmlWriter(buf, "item")
		if err != nil {
			return err
		}
	}
	buf.Write([]byte("</items>"))
	return nil
}
