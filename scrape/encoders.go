package scrape

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
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
		json.NewEncoder(&buf).Encode(results.Output)
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
	allBlocks := results.AllBlocks()
	err := encodeCSV(e.partNames, allBlocks, e.comma, w)
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

func intArrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}

func floatArrayToString(a []float64, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
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
			case int:
				r[i] = strconv.FormatInt(int64(v), 10)
			case []int:
				r[i] = intArrayToString(v, ";")
			case []float64:
				r[i] = floatArrayToString(v, ";")
			case float64:
				r[i] = strconv.FormatFloat(v, 'f', -1, 64)
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
	buf.Write([]byte("<root>"))
	for _, elem := range blocks {
		nm := make(map[string]interface{})
		for key, value := range elem {
			out := []string{}
			//[]int and []float slices should be passed as []string
			switch v := value.(type) {
			case []int:
				for _, i := range v {
					out = append(out, strconv.Itoa(i))
				}
				nm[key] = out
				elem = nm
			case []float64:
				for _, i := range v {
					out = append(out, strconv.FormatFloat(i, 'f', -1, 64))
				}
				nm[key] = out
				elem = nm
			}
		} 
		m := mxj.Map(elem)
		//err := m.XmlIndentWriter(&buf, "", "  ", "object")
		err := m.XmlWriter(buf, "element")
		if err != nil {
			return err
		}
	}
	buf.Write([]byte("</root>"))
	return nil
}
