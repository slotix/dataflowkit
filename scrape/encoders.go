package scrape

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/storage"
	"github.com/spf13/viper"
)

// EncodeToFile save parsed data to specified file.
func EncodeToFile(e *encoder, ext string, payloadMD5 string, blockMap ...*map[int][]int) ([]byte, error) {
	path := viper.GetString("RESULTS_DIR")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0700)
	}
	sFileName := viper.GetString("RESULTS_DIR") + "/" + payloadMD5 + "_" + time.Now().Format("2006-01-02_15:04") + "." + ext
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
	var keys *map[int][]int
	if len(blockMap) > 0 {
		keys = blockMap[0]
	}
	w := bufio.NewWriter(fo)
	(*e).encode(w, payloadMD5, keys)
	return []byte(sFileName), nil
}

// func EncodeToByteArray(e *encoder, payloadMD5 string, blockMap ...*map[int][]int) ([]byte, error) {
// 	result := ""
// 	buf := bytes.NewBufferString(result)
// 	w := bufio.NewWriter(buf)
// 	var keys *map[int][]int
// 	if len(blockMap) > 0 {
// 		keys = blockMap[0]
// 	}
// 	(*e).encode(w, payloadMD5, keys)
// 	return buf.Bytes(), nil
// }

type encoder interface {
	encode(w *bufio.Writer, payloadMD5 string, keys *map[int][]int) error
}

// CSVEncoder transforms parsed data to CSV format.
type CSVEncoder struct {
	partNames []string
	comma     string
}

// JSONEncoder transforms parsed data to JSON format.
type JSONEncoder struct {
	//	paginateResults bool
}

// XMLEncoder transforms parsed data to XML format.
type XMLEncoder struct {
}

func (e JSONEncoder) encode(w *bufio.Writer, payloadMD5 string, keys *map[int][]int) error {
	storageType := viper.GetString("STORAGE_TYPE")
	s := storage.NewStore(storageType)
	// make a write buffer
	// if e.paginateResults {
	// 	w.WriteString("[")
	// }
	w.WriteString("[")

	reader := newStorageReader(&s, payloadMD5, keys)
	writeComma := false
	for {
		block, err := reader.Read()
		if err != nil {
			if err.Error() == errs.EOF {
				w.WriteString("]")
				break
			} else if err.Error() == errs.NextPage {
				//next page
				// if e.paginateResults {
				// 	w.WriteString("],[")
				// }
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
	// if e.paginateResults {
	// 	w.WriteString("]")
	// }
	s.Close()
	return w.Flush()
}

type storageResultReader struct {
	storage    *storage.Store
	payloadMD5 string
	page       int
	keys       []int
	block      int
	payloadMap map[int][]int
}

func newStorageReader(store *storage.Store, md5Hash string, extractPageKeys *map[int][]int) *storageResultReader {
	reader := &storageResultReader{
		storage:    store,
		payloadMD5: md5Hash,
		payloadMap: make(map[int][]int),
		block:      0,
		page:       0,
	}
	if extractPageKeys != nil {
		reader.payloadMap = *extractPageKeys
	}
	reader.init()
	return reader
}

// have return error
func (r *storageResultReader) init() {
	r.page = 0
	r.block = 0
	if len(r.payloadMap) == 0 {
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
	}
	for k := range r.payloadMap {
		r.keys = append(r.keys, k)
	}
}

// func (r *storageResultReader) initManualKeys(blocks []int) {
// 	r.keys = blocks
// }

func (r *storageResultReader) Read() (map[string]interface{}, error) {
	//blockMap := make(map[string]interface{})
	var err error
	var nextPage bool
	if r.block >= len(r.payloadMap[r.keys[r.page]]) {
		if r.page+1 < len(r.keys) {
			//achieve next page
			r.page++
			r.block = 0
			nextPage = true
		} else {
			//achieve EOF
			return nil, &errs.ErrStorageResult{Err: errs.EOF}
		}
	}
	blockMap, err := r.getValue()
	if err != nil {
		// have to try get next block value
		r.block++
		return nil, err
	}
	for field, value := range blockMap {
		if strings.Contains(field, "details") {
			details := []map[string]interface{}{}
			detailsReader := newStorageReader(r.storage, value.(string), nil)
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
				details = append(details, detailsBlock)
				// we are just breaking here because we got all details recursively
				// if we will continue to read storage, next iteration will returns EOF
				// just to save a time break loop manually
				//break
			}
			if len(details) > 0 {
				if len(details) == 1 {
					blockMap[field] = details[0]
				} else {
					blockMap[field] = details
				}
			}
		}
	}
	r.block++
	if nextPage {
		err = &errs.ErrStorageResult{Err: errs.NextPage}
	}
	return blockMap, err
}

func (r *storageResultReader) getValue() (map[string]interface{}, error) {
	key := fmt.Sprintf("%s-%d-%d", r.payloadMD5, r.page, r.payloadMap[r.page][r.block])
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

func (e CSVEncoder) encode(w *bufio.Writer, payloadMD5 string, keys *map[int][]int) error {
	storageType := viper.GetString("STORAGE_TYPE")
	s := storage.NewStore(storageType)

	//write csv headers
	sString := ""
	for _, headerName := range e.partNames {
		sString += fmt.Sprintf("%s,", headerName)
	}
	sString = strings.TrimSuffix(sString, ",") + "\n"
	_, err := w.WriteString(sString)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}

	reader := newStorageReader(&s, payloadMD5, keys)

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
			sString += e.formatFieldValue(block[fieldName])
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
	return w.Flush()
}

func (e CSVEncoder) formatFieldValue(v interface{}) string {
	formatedString := ""
	switch v.(type) {
	case string:
		formatedString = v.(string)
	case []string:
		formatedString = strings.Join(v.([]string), ";")
	case int:
		formatedString = strconv.FormatInt(int64(v.(int)), 10)
	case []int:
		formatedString = intArrayToString(v.([]int), ";")
	case []float64:
		formatedString = floatArrayToString(v.([]float64), ";")
	case float64:
		formatedString = strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case nil:
		formatedString = ""
	case []interface{}:
		values := make([]string, len(v.(string)))
		for i, value := range v.(string) {
			values[i] = fmt.Sprint(value)
		}
		formatedString = strings.Join(values, ";")
	}
	return fmt.Sprintf("%s,", formatedString)
}

func (e XMLEncoder) encode(w *bufio.Writer, payloadMD5 string, keys *map[int][]int) error {
	storageType := viper.GetString("STORAGE_TYPE")
	s := storage.NewStore(storageType)
	//write xml headers
	_, err := w.WriteString(`<?xml version="1.0" encoding="UTF-8"?><root>`)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	reader := newStorageReader(&s, payloadMD5, keys)

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
	return w.Flush()
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
			v, ok := value.(string)
			if ok {
				xml.Escape(w, []byte(v))
			} else {
				for i, val := range value.([]interface{}) {
					v, ok = val.(string)
					if ok {
						xml.Escape(w, []byte(v))
						if i < len(value.([]interface{}))-1 {
							w.Write([]byte(";"))
						}
					}
				}
			}
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
// func encodeCSV(header []string, rows []map[string]interface{}, comma string, w *csv.Writer) error {
// 	if comma == "" {
// 		comma = ","
// 	}
// 	w.Comma = rune(comma[0])
// 	//Add Header string to csv or no
// 	if len(header) > 0 {
// 		if err := w.Write(header); err != nil {
// 			return err
// 		}
// 	}
// 	r := make([]string, len(header))
// 	for _, row := range rows {
// 		for i, column := range header {
// 			switch v := row[column].(type) {
// 			case string:
// 				r[i] = v
// 			case []string:
// 				r[i] = strings.Join(v, ";")
// 			case int:
// 				r[i] = strconv.FormatInt(int64(v), 10)
// 			case []int:
// 				r[i] = intArrayToString(v, ";")
// 			case []float64:
// 				r[i] = floatArrayToString(v, ";")
// 			case float64:
// 				r[i] = strconv.FormatFloat(v, 'f', -1, 64)
// 			case nil:
// 				r[i] = ""
// 			}
// 		}
// 		if err := w.Write(r); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

//encodeXML writes data blocks to XML file.
// func encodeXML(blocks []map[string]interface{}, buf *bytes.Buffer) error {
// 	mxj.XMLEscapeChars(true)
// 	//write header to xml
// 	buf.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
// 	buf.Write([]byte("<root>"))
// 	for _, elem := range blocks {
// 		nm := make(map[string]interface{})
// 		for key, value := range elem {
// 			out := []string{}
// 			//[]int and []float slices should be passed as []string
// 			switch v := value.(type) {
// 			case []int:
// 				for _, i := range v {
// 					out = append(out, strconv.Itoa(i))
// 				}
// 				nm[key] = out
// 				elem = nm
// 			case []float64:
// 				for _, i := range v {
// 					out = append(out, strconv.FormatFloat(i, 'f', -1, 64))
// 				}
// 				nm[key] = out
// 				elem = nm
// 			}
// 		}
// 		m := mxj.Map(elem)
// 		//err := m.XmlIndentWriter(&buf, "", "  ", "object")
// 		err := m.XmlWriter(buf, "element")
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	buf.Write([]byte("</root>"))
// 	return nil
// }
