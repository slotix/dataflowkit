package scrape

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/clbanning/mxj"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/storage"
	"github.com/spf13/viper"
)

//bug: xml is not correct if there are details in payload
//

type encoder interface {
	Encode(*Results) (io.ReadCloser, error)
	EncodeFromStorage(payloadMD5 string) (io.ReadCloser, error)
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

func (e JSONEncoder) EncodeFromStorage(payloadMD5 string) (io.ReadCloser, error) {
	storageType := viper.GetString("STORAGE_TYPE")
	s := storage.NewStore(storageType)
	// open output file
	sFileName := payloadMD5 + "_" + time.Now().Format("2006-01-02_15:04") + ".json"
	fo, err := os.OpenFile(sFileName, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		return nil, err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	// make a write buffer
	w := bufio.NewWriter(fo)
	if e.paginateResults {
		w.WriteString("[")
	}
	w.WriteString("[")

	reader := newStorageReader(&s, payloadMD5)
	writeComma := false
	for {
		block, err := reader.Read()
		if err != nil {
			if err.Error() == errs.EOF {
				w.WriteString("]")
				break
			} else if err.Error() == errs.NextPage {
				//next page
				if e.paginateResults {
					w.WriteString("],[")
				}
			} else {
				logger.Error(err)
				continue
			}
		}
		if writeComma {
			w.WriteString(",")
		}
		blockJSON, err := json.Marshal(block)
		if err != nil {
			logger.Error(err)
		}
		if !writeComma {
			writeComma = !writeComma
		}
		w.Write(blockJSON)
	}
	if e.paginateResults {
		w.WriteString("]")
	}
	s.Close()
	if err = w.Flush(); err != nil {
		return nil, err
	}
	readCloser := ioutil.NopCloser(strings.NewReader(sFileName))
	return readCloser, nil
}

type storageResultReader struct {
	storage    *storage.Store
	payloadMD5 string
	page       int
	keys       []int
	block      int
	payloadMap map[int][]int
}

func newStorageReader(store *storage.Store, md5Hash string) *storageResultReader {
	reader := &storageResultReader{
		storage:    store,
		payloadMD5: md5Hash,
		payloadMap: make(map[int][]int),
		block:      0,
		page:       0,
	}
	reader.init()
	return reader
}

// have return error
func (r *storageResultReader) init() {
	keysJSON, err := (*r.storage).Read(storage.Record{
		Type: storage.INTERMEDIATE,
		Key:  r.payloadMD5,
	})
	if err != nil {
		logger.Error(err)
	}
	err = json.Unmarshal(keysJSON, &r.payloadMap)
	if err != nil {
		logger.Error(err)
	}

	for k := range r.payloadMap {
		r.keys = append(r.keys, k)
	}
	r.page = 0
	r.block = 0
}

func (r *storageResultReader) Read() (map[string]interface{}, error) {
	blockMap := make(map[string]interface{})
	var err error
	if r.block == len(r.payloadMap[r.keys[r.page]]) {
		if r.page+1 < len(r.keys) {
			//achieve next page
			r.page++
			r.block = 0
			err = &errs.ErrStorageResult{Err: errs.NextPage}
		} else {
			//achieve EOF
			return nil, &errs.ErrStorageResult{Err: errs.EOF}
		}
	}
	blockMap, err = r.getValue()
	if err != nil {
		// have to try get next block value
		r.block++
		return nil, err
	}
	for field, value := range blockMap {
		if strings.Contains(field, "details") {
			details := map[string]interface{}{}
			detailsReader := newStorageReader(r.storage, value.(string))
			for {
				detailsBlock, detailsErr := detailsReader.Read()
				if detailsErr != nil {
					if detailsErr.Error() == errs.NextPage {

					} else if detailsErr.Error() == errs.EOF {
						break
					} else {
						// in a case of details "no such file or directory" error means, that
						// detail's selector(s) has not be found in a block
						// these can happens when there are few coresponding blocks within a page
						// but only some of them contains wanted selector(s)
						// so just go ahead
						continue
					}
				}
				details = detailsBlock
				// we are just breaking here because we got all details recursively
				// if we will continue to read storage, next iteration will returns EOF
				// just to save a time break loop manualy
				break
			}
			if len(details) > 0 {
				blockMap[field] = details
			}
		}
	}
	r.block++
	return blockMap, err
}

func (r *storageResultReader) getValue() (map[string]interface{}, error) {
	key := fmt.Sprintf("%s-%d-%d", r.payloadMD5, r.page, r.block)
	blockJSON, err := (*r.storage).Read(storage.Record{
		Type: storage.INTERMEDIATE,
		Key:  key,
	})

	if err != nil {
		return nil, err //&errs.ErrStorageResult{Err: fmt.Sprintf(errs.NoKey, key)}
	}
	blockMap := make(map[string]interface{})
	err = json.Unmarshal(blockJSON, &blockMap)
	if err != nil {
		return nil, err
	}
	return blockMap, nil
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

func (e CSVEncoder) EncodeFromStorage(payloadMD5 string) (io.ReadCloser, error) {
	storageType:= viper.GetString("STORAGE_TYPE")
	s := storage.NewStore(storageType)
	// open output file
	sFileName := payloadMD5 + "_" + time.Now().Format("2006-01-02_15:04") + ".csv"
	fo, err := os.OpenFile(sFileName, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		return nil, err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	// make a write buffer
	w := bufio.NewWriter(fo)

	//write csv headers
	sString := ""
	for _, headerName := range e.partNames {
		sString += fmt.Sprintf("%s,", headerName)
	}
	sString = strings.TrimSuffix(sString, ",") + "\n"
	_, err = w.WriteString(sString)
	if err != nil {
		logger.Error(err)
	}
	err = w.Flush()
	if err != nil {
		logger.Error(err)
	}

	reader := newStorageReader(&s, payloadMD5)

	for {
		block, err := reader.Read()
		if err != nil {
			if err.Error() == errs.EOF {
				break
			} else if err.Error() == errs.NextPage {
				//next page
			} else {
				logger.Error(err)
				//we have to continue 'cause we still have other records
				continue
			}
		}
		sString = ""
		for _, fieldName := range e.partNames {
			formatedString := ""
			switch v := block[fieldName].(type) {
			case string:
				formatedString = v
			case []string:
				formatedString = strings.Join(v, ";")
			case int:
				formatedString = strconv.FormatInt(int64(v), 10)
			case []int:
				formatedString = intArrayToString(v, ";")
			case []float64:
				formatedString = floatArrayToString(v, ";")
			case float64:
				formatedString = strconv.FormatFloat(v, 'f', -1, 64)
			case nil:
				formatedString = ""
			case []interface{}:
				values := make([]string, len(v))
				for i, value := range v {
					values[i] = fmt.Sprint(value)
				}
				formatedString = strings.Join(values, ";")
			}
			sString += fmt.Sprintf("%s,", formatedString)
		}
		sString = strings.TrimSuffix(sString, ",") + "\n"
		_, err = w.WriteString(sString)
		if err != nil {
			logger.Error(err)
		}
		err = w.Flush()
		if err != nil {
			logger.Error(err)
		}
	}

	s.Close()

	if err = w.Flush(); err != nil {
		panic(err)
	}
	readCloser := ioutil.NopCloser(strings.NewReader(sFileName))
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

func (e XMLEncoder) EncodeFromStorage(payloadMD5 string) (io.ReadCloser, error) {
	storageType := viper.GetString("STORAGE_TYPE")
	s := storage.NewStore(storageType)
	// open output file
	sFileName := payloadMD5 + "_" + time.Now().Format("2006-01-02_15:04") + ".xml"
	fo, err := os.OpenFile(sFileName, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		return nil, err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	// make a write buffer
	w := bufio.NewWriter(fo)

	//write xml headers
	_, err = w.WriteString(`<?xml version="1.0" encoding="UTF-8"?><root>`)
	if err != nil {
		logger.Error(err)
	}
	err = w.Flush()
	if err != nil {
		logger.Error(err)
	}

	reader := newStorageReader(&s, payloadMD5)

	for {
		block, err := reader.Read()
		if err != nil {
			if err.Error() == errs.EOF {
				break
			} else if err.Error() == errs.NextPage {
				//next page
			} else {
				logger.Error(err)
				//we have to continue 'cause we still have other records
				continue
			}
		}
		e.writeXML(w, &block)
		err = w.Flush()
		if err != nil {
			logger.Error(err)
		}
	}

	s.Close()

	w.WriteString("</root>")
	if err = w.Flush(); err != nil {
		panic(err)
	}
	readCloser := ioutil.NopCloser(strings.NewReader(sFileName))
	return readCloser, nil
}

func (e XMLEncoder) writeXML(w io.Writer, block *map[string]interface{}) {
	for field, value := range *block {
		if strings.Contains(field, "details") {
			v := value.(map[string]interface{})
			w.Write([]byte(fmt.Sprintf("<%s>", field)))
			e.writeXML(w, &v)
			w.Write([]byte(fmt.Sprintf("</%s>", field)))
		} else {
			nodeName := fmt.Sprintf("<%s>", field)
			w.Write([]byte(nodeName))
			// have to escape predefined entities to obtain valid xml
			xml.Escape(w, []byte(value.(string)))
			nodeName = fmt.Sprintf("</%s>", field)
			w.Write([]byte(nodeName))
		}
	}
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
