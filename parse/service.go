package parse

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/slotix/dataflowkit/scrape"
)

// Define service interface
type Service interface {
	Parse(scrape.Payload) (io.ReadCloser, error)
}

// Implement service with empty struct
type ParseService struct {
}

type ServiceMiddleware func(Service) Service

//Parse calls Fetcher which downloads web page content for parsing
func (ps ParseService) Parse(p scrape.Payload) (io.ReadCloser, error) {
	task, err := scrape.NewTask(p)
	if err != nil {
		return nil, err
	}
	//scrape request and return results.
	err = scrape.Scrape(task)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	opts := task.Scraper.Opts
	switch opts.Format {
	case "json":
		if opts.PaginateResults {
			json.NewEncoder(&buf).Encode(task.Results)
		} else {
			json.NewEncoder(&buf).Encode(task.Results.AllBlocks())
		}
	case "csv":
		/*
			includeHeader := true
			w := csv.NewWriter(&buf)
			for i, page := range results.Results {
				if i != 0 {
					includeHeader = false
				}
				err = encodeCSV(names, includeHeader, page, ",", w)
				if err != nil {
					logger.Println(err)
				}
			}
			w.Flush()
		*/
		w := csv.NewWriter(&buf)

		err = encodeCSV(task.Scraper.PartNames(), task.Results.AllBlocks(), ",", w)
		w.Flush()
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
				logger.Println(err)
			}
	*/
	case "xml":
		err = encodeXML(task.Results.AllBlocks(), &buf)
		if err != nil {
			return nil, err
		}
	}
	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}
