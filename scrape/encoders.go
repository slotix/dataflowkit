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

func newEncoder(s Task) (e encoder) {
	switch strings.ToLower(s.Scrapers[0].Opts.Format) {
	case "csv":
		e = CSVEncoder{
			comma:     ",",
			partNames: s.Scrapers[0].partNames(),
			results:   &s.Results,
		}
		return
	case "json":
		e = JSONEncoder{
			paginateResults: s.Scrapers[0].Opts.PaginateResults,
			results:         &s.Results,
		}
		return
	case "xml":
		e = XMLEncoder{
			results: &s.Results,
		}
		return e
	default:
		return nil
	}
}

type encoder interface {
	Encode() (io.ReadCloser, error)
	Format() string
}

// CSVEncoder transforms parsed data to CSV format.
type CSVEncoder struct {
	partNames []string
	comma     string
	results   *Results
}

// JSONEncoder transforms parsed data to JSON format.
type JSONEncoder struct {
	paginateResults bool
	results         *Results
}

// XMLEncoder transforms parsed data to XML format.
type XMLEncoder struct {
	results *Results
}

//Encode method implementation for JSONEncoder
func (e JSONEncoder) Encode() (io.ReadCloser, error) {
	var buf bytes.Buffer
	if e.paginateResults {
		json.NewEncoder(&buf).Encode(e.results)
	} else {
		json.NewEncoder(&buf).Encode(e.results.AllBlocks())
	}
	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}

//Format returns format of JSONEncoder
func (JSONEncoder) Format() string {
	return "JSON"
}

//Encode method implementation for CSVEncoder
func (e CSVEncoder) Encode() (io.ReadCloser, error) {
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

	err := encodeCSV(e.partNames, e.results.AllBlocks(), e.comma, w)
	if err != nil {
		return nil, err
	}
	w.Flush()
	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}

//Format returns format of CSVEncoder
func (CSVEncoder) Format() string {
	return "CSV"
}

//Encode method implementation for XMLEncoder
func (e XMLEncoder) Encode() (io.ReadCloser, error) {
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
	err := encodeXML(e.results.AllBlocks(), &buf)
	if err != nil {
		return nil, err
	}
	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}

//Format returns format of XMLEncoder
func (XMLEncoder) Format() string {
	return "XML"
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
