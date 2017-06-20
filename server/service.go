package server

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"fmt"

	"github.com/clbanning/mxj"
	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/helpers"
	"github.com/slotix/dataflowkit/paginate"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/splash"
)

// ParseService provides operations on strings.
type ParseService interface {
	//	GetResponse(req splash.Request) (*splash.Response, error)
	Fetch(req splash.Request) (interface{}, error)
	ParseData(payload []byte) (io.ReadCloser, error)
	//	CheckServices() (status map[string]string)
}

type parseService struct {
}

//Fetch returns splash.Request
func (parseService) Fetch(req splash.Request) (interface{}, error) {
	//logger.Println(req)
	fetcher, err := scrape.NewSplashFetcher()

	if err != nil {
		logger.Println(err)
	}
	res, err := fetcher.Fetch(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type Extractor struct {
	Type   string      `json:"type"`
	Params interface{} `json:"params"`
}

type field struct {
	Name      string    `json:"name" validate:"required"`
	Selector  string    `json:"selector" validate:"required"`
	Count     int       `json:"count"`
	Details   Payload   `json:"-" validate:"-"`
	Extractor Extractor `json:"extractor"`
}

type paginator struct {
	Selector  string `json:"selector"`
	Attribute string `json:"attr"`
	MaxPages  int    `json:"maxPages"`
}

type Payload struct {
	Name             string    `json:"name" xml:"name" validate:"required"`
	URL              string    `json:"url" xml:"url" validate:"required"`
	Fields           []field   `json:"fields" validate:"gt=0"` //number of fields
	Paginator        paginator `json:"paginator"`
	PayloadMD5       []byte    `json:"payloadMD5"`
	Format           string    `json:"format"`
	PaginatedResults bool      `json:"paginatedResults"`
}

//NewParser initializes new Parser struct
func NewPayload(payload []byte) (Payload, error) {
	var p Payload
	err := json.Unmarshal(payload, &p)
	//err := p.UnmarshalJSON(payload)
	if err != nil {
		return p, err
	}
	if p.Format == "" {
		p.Format = "json"
	}
	p.PayloadMD5 = helpers.GenerateMD5(payload)
	return p, nil
}

func (p Payload) payloadToScrapeConfig() (config *scrape.ScrapeConfig, err error) {

	if err != nil {
		return nil, err
	}
	fetcher, err := scrape.NewSplashFetcher()
	if err != nil {
		logger.Println(err)
	}
	pieces := []scrape.Piece{}
	//pl := p.Payloads[0]
	selectors := []string{}
	names := []string{}
	for _, f := range p.Fields {
		//var extractor scrape.PieceExtractor
		params := make(map[string]interface{})
		if f.Extractor.Params != nil {
			params = f.Extractor.Params.(map[string]interface{})
		}
		switch eType := f.Extractor.Type; eType {
		//For Link type by default Two pieces with different Text and Attr="href" extractors will be added for field selector.
		case "link":
			t := &extract.Text{}
			if params != nil {
				err := extract.FillStruct(params, t)
				if err != nil {
					logger.Println(err)
				}
			}
			fName := fmt.Sprintf("%s_text", f.Name)
			pieces = append(pieces, scrape.Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: t,
			})
			names = append(names, fName)

			a := &extract.Attr{Attr: "href"}
			if params != nil {
				err := extract.FillStruct(params, a)
				if err != nil {
					logger.Println(err)
				}
			}
			fName = fmt.Sprintf("%s_link", f.Name)
			pieces = append(pieces, scrape.Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: a,
			})
			names = append(names, fName)
			//Add selector just one time for link type
			selectors = append(selectors, f.Selector)
		//For image type by default Two pieces with different Attr="src" and Attr="alt" extractors will be added for field selector.
		case "image":
			a := &extract.Attr{Attr: "src"}
			if params != nil {
				err := extract.FillStruct(params, a)
				if err != nil {
					logger.Println(err)
				}
			}
			fName := fmt.Sprintf("%s_src", f.Name)
			pieces = append(pieces, scrape.Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: a,
			})
			names = append(names, fName)

			a = &extract.Attr{Attr: "alt"}
			if params != nil {
				err := extract.FillStruct(params, a)
				if err != nil {
					logger.Println(err)
				}
			}
			fName = fmt.Sprintf("%s_alt", f.Name)
			pieces = append(pieces, scrape.Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: a,
			})
			names = append(names, fName)
			//Add selector just one time for link type
			selectors = append(selectors, f.Selector)

		default:
			var e extract.PieceExtractor
			switch eType {
			case "text":
				e = &extract.Text{}
			case "attr":
				e = &extract.Attr{}
			case "regex":
				r := &extract.Regex{}
				regExp := params["regexp"]
				r.Regex = regexp.MustCompile(regExp.(string))
				e = r
			}
			//extractor, err := extract.FillParams(f.Extractor.Type, params)
			//err := e.FillParams(params)
			if params != nil {
				err := extract.FillStruct(params, e)
				if err != nil {
					logger.Println(err)
				}
			}
			pieces = append(pieces, scrape.Piece{
				Name:      f.Name,
				Selector:  f.Selector,
				Extractor: e,
			})
			selectors = append(selectors, f.Selector)
			names = append(names, f.Name)
		}
	}

	paginator := p.Paginator
	config = &scrape.ScrapeConfig{
		Fetcher: fetcher,
		//DividePage: scrape.DividePageBySelector(".p"),
		DividePage: scrape.DividePageByIntersection(selectors),
		Pieces:     pieces,
		//Paginator: paginate.BySelector(".next", "href"),
		CSVHeader: names,
		Paginator: paginate.BySelector(paginator.Selector, paginator.Attribute),
		Opts: scrape.ScrapeOptions{
			MaxPages:         paginator.MaxPages,
			Format:           p.Format,
			PaginatedResults: p.PaginatedResults},
	}
	return
}

func (parseService) ParseData(payload []byte) (io.ReadCloser, error) {
	p, err := NewPayload(payload)
	if err != nil {
		return nil, err
	}
	config, err := p.payloadToScrapeConfig()
	if err != nil {
		return nil, err
	}
	scraper, err := scrape.New(config)
	if err != nil {
		return nil, err
	}
	req := splash.Request{URL: p.URL}
	results, err := scraper.ScrapeWithOpts(req, config.Opts)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	switch config.Opts.Format {
	case "json":
		if config.Opts.PaginatedResults {
			json.NewEncoder(&buf).Encode(results)
		} else {
			json.NewEncoder(&buf).Encode(results.AllBlocks())
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
		err = encodeCSV(config.CSVHeader, true, results.AllBlocks(), ",", w)
		w.Flush()
	/*
		case "xmlviajson":
			var jbuf bytes.Buffer
			if config.Opts.PaginatedResults {
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
		err = encodeXML(results.AllBlocks(), &buf)
		if err != nil{
			return nil, err
		}
		/*
		mxj.XMLEscapeChars(true)
		//write header to xml
		buf.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
		buf.Write([]byte("<doc>"))
		for _, piece := range results.AllBlocks() {
			m := mxj.Map(piece)
			//err := m.XmlIndentWriter(&buf, "", "  ", "object")
			err := m.XmlWriter(&buf, "object")
			if err != nil {
				logger.Println(err)
			}
		}
		buf.Write([]byte("</doc>"))
		*/
	}

	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}

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

func encodeXML(blocks []map[string]interface{}, buf *bytes.Buffer) error {
	mxj.XMLEscapeChars(true)
	//write header to xml
	buf.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
	buf.Write([]byte("<doc>"))
	for _, piece := range blocks {
		m := mxj.Map(piece)
		//err := m.XmlIndentWriter(&buf, "", "  ", "object")
		err := m.XmlWriter(buf, "object")
		if err != nil {
			return err
		}
	}
	buf.Write([]byte("</doc>"))
	return nil
}

/*
func (parseService) GetResponse(req splash.Request) (*splash.Response, error) {
	splashURL, err := splash.NewSplashConn(req)
	response, err := splash.GetResponse(splashURL)
	return response, err
}
*/

/*
func (parseService) Fetch(req splash.Request) (io.ReadCloser, error) {
	logger.Println(req)
	splashURL, err := splash.NewSplashConn(req)
	content, err := splash.Fetch(splashURL)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (parseService) ParseData_old(payload []byte) (io.ReadCloser, error) {
	p, err := parser.NewParser(payload)
	if err != nil {
		return nil, err
	}
	res, err := p.MarshalData()
	if err != nil {
		logger.Println(res, err)
		return nil, err
	}
	return res, nil
}
*/
//func (parseService) CheckServices() (status map[string]string) {
//	return CheckServices() //, allAlive
//}

// ServiceMiddleware is a chainable behavior modifier for ParseService.
type ServiceMiddleware func(ParseService) ParseService
