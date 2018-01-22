package scrape

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/clbanning/mxj"
)

//bug: xml is not correct if there are details in payload
//

type encoder interface {
	Encode(*Results) (io.ReadCloser, error)
}

// CSVEncoder transforms parsed data to CSV format.
type CSVEncoder struct {
	partNames []string
	comma     string
}

// JSONEncoder transforms parsed data to JSON format.
type JSONEncoder struct {
	paginateResults bool
}

// XMLEncoder transforms parsed data to XML format.
type XMLEncoder struct {
}

//Encode method implementation for JSONEncoder
func (e JSONEncoder) Encode(results *Results) (io.ReadCloser, error) {
	var buf bytes.Buffer
	if e.paginateResults {
		json.NewEncoder(&buf).Encode(results)
	} else {
		json.NewEncoder(&buf).Encode(results.AllBlocks())
	}
	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}

//Encode method implementation for CSVEncoder
func (e CSVEncoder) Encode(results *Results) (io.ReadCloser, error) {
	var buf bytes.Buffer
	/*
		includeHeader := true
		w := csv.NewWriter(&buf)
		for i, page := range results.Results {
			if i != 0 {
				includeHeader = false
			}
			err = encodeCSV(names, includeHeader, page, ",", w)
			if err != nil {
				logger.Error(err)
			}
		}
		w.Flush()
	*/
	w := csv.NewWriter(&buf)

	err := encodeCSV(e.partNames, results.AllBlocks(), e.comma, w)
	if err != nil {
		return nil, err
	}
	w.Flush()
	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}



//Encode method implementation for XMLEncoder
func (e XMLEncoder) Encode(results *Results) (io.ReadCloser, error) {
	/*
		case "xmlviajson":
			var jbuf bytes.Buffer
			if config.Opts.PaginateResults {
				json.NewEncoder(&jbuf).Encode(results)
			} else {
				json.NewEncoder(&jbuf).Encode(results.AllBlocks())
			}
			//var buf bytes.Buffer
			m, err := mxj.NewMapJson(jbuf.Bytes())
			err = m.XmlIndentWriter(&buf, "", "  ")
			if err != nil {
				logger.Error(err)
			}
	*/
	var buf bytes.Buffer
	err := encodeXML(results.AllBlocks(), &buf)
	if err != nil {
		return nil, err
	}
	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}


//encodeCSV writes data to w *csv.Writer.
//header represent an array of fields for csv.
//rows store csv records to be written.
//comma is a separator between record fields. Default value is ","
func encodeCSV(header []string, rows []map[string]interface{}, comma string, w *csv.Writer) error {
	if comma == "" {
		comma = ","
	}
	w.Comma = rune(comma[0])
	//Add Header string to csv or no
	if len(header) > 0 {
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

//encodeXML writes data blocks to XML file.
func encodeXML(blocks []map[string]interface{}, buf *bytes.Buffer) error {
	mxj.XMLEscapeChars(true)
	//write header to xml
	buf.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
	buf.Write([]byte("<items>"))
	for _, elem := range blocks {
		m := mxj.Map(elem)
		//err := m.XmlIndentWriter(&buf, "", "  ", "object")
		err := m.XmlWriter(buf, "item")
		if err != nil {
			return err
		}
	}
	buf.Write([]byte("</items>"))
	return nil
}
