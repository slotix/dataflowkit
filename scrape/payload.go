package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/paginate"
	"github.com/slotix/dataflowkit/splash"
)

//http://choly.ca/post/go-json-marshalling/
//UnmarshalJSON convert headers to http.Header type
//Unmarshal request as splash.Request
func (p *Payload) UnmarshalJSON(data []byte) error {
	type Alias Payload
	aux := &struct {
		Request interface{} `json:"request"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	logger.Println(aux.Request)
	splashRequest := splash.Request{}
	//err := FillStruct(aux.Request.(map[string]interface{}), splashRequest)
	err := FillStruct(aux.Request.(map[string]interface{}), &splashRequest)
	if err != nil {
		return err
	}
	p.Request = splashRequest
	//logger.Println(splashRequest)
	return nil
}

//NewParser initializes new Parser struct
func NewPayload(payload []byte) (Payload, error) {
	var p Payload
	err := json.Unmarshal(payload, &p)
	
	if err != nil {
		return p, err
	}
	if p.Format == "" {
		p.Format = DefaultOptions.Format
	}
	if p.RetryTimes == 0 {
		p.RetryTimes = DefaultOptions.RetryTimes
	}
	if p.FetchDelay == 0 {
		p.FetchDelay = DefaultOptions.FetchDelay
	}
	if p.RandomizeFetchDelay == nil {
		p.RandomizeFetchDelay = &DefaultOptions.RandomizeFetchDelay
	}
	if p.PaginateResults == nil {
		p.PaginateResults = &DefaultOptions.PaginateResults
	}

	p.PayloadMD5 = GenerateMD5(payload)
	return p, nil
}

func (p Payload) PayloadToScrapeConfig() (config *ScrapeConfig, err error) {
	fetcher, err := NewSplashFetcher()
	//fetcher, err := NewHttpClientFetcher()
	
	if err != nil {
		logger.Println(err)
	}
	pieces := []Piece{}
	selectors := []string{}
	names := []string{}
	for _, f := range p.Fields {
		//var extractor scrape.PieceExtractor
		params := make(map[string]interface{})
		if f.Extractor.Params != nil {
			params = f.Extractor.Params.(map[string]interface{})
		}
		switch eType := f.Extractor.Type; eType {
		//For Link type Two pieces as pair Text and Attr{Attr:"href"} extractors are added.
		case "link":
			t := &extract.Text{}
			if params != nil {
				err := FillStruct(params, t)
				if err != nil {
					logger.Println(err)
				}
			}
			fName := fmt.Sprintf("%s_text", f.Name)
			pieces = append(pieces, Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: t,
			})
			names = append(names, fName)

			a := &extract.Attr{Attr: "href"}
			if params != nil {
				err := FillStruct(params, a)
				if err != nil {
					logger.Println(err)
				}
			}
			fName = fmt.Sprintf("%s_link", f.Name)
			pieces = append(pieces, Piece{
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
				err := FillStruct(params, a)
				if err != nil {
					logger.Println(err)
				}
			}
			fName := fmt.Sprintf("%s_src", f.Name)
			pieces = append(pieces, Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: a,
			})
			names = append(names, fName)

			a = &extract.Attr{Attr: "alt"}
			if params != nil {
				err := FillStruct(params, a)
				if err != nil {
					logger.Println(err)
				}
			}
			fName = fmt.Sprintf("%s_alt", f.Name)
			pieces = append(pieces, Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: a,
			})
			names = append(names, fName)
			//Add selector just one time for image type
			selectors = append(selectors, f.Selector)

		default:
			var e extract.PieceExtractor
			switch eType {
			case "const":
				//	c := &extract.Const{Val: params["value"]}
				//	e = c
				e = &extract.Const{}
			case "count":
				e = &extract.Count{}
			case "text":
				e = &extract.Text{}
			case "multipleText":
				e = &extract.MultipleText{}
			case "html":
				e = &extract.Html{}
			case "outerHtml":
				e = &extract.OuterHtml{}
			case "attr":
				e = &extract.Attr{}
			case "regex":
				r := &extract.Regex{}
				regExp := params["regexp"]
				r.Regex = regexp.MustCompile(regExp.(string))
				e = r
			}

			if params != nil {
				err := FillStruct(params, e)
				if err != nil {
					logger.Println(err)
				}
			}
			pieces = append(pieces, Piece{
				Name:      f.Name,
				Selector:  f.Selector,
				Extractor: e,
			})
			selectors = append(selectors, f.Selector)
			names = append(names, f.Name)
		}
	}

	paginator := p.Paginator
	config = &ScrapeConfig{
		Fetcher: fetcher,
		//DividePage: scrape.DividePageBySelector(".p"),
		DividePage: DividePageByIntersection(selectors),
		Pieces:     pieces,
		//Paginator: paginate.BySelector(".next", "href"),
		CSVHeader: names,
		Paginator: paginate.BySelector(paginator.Selector, paginator.Attribute),
		Opts: ScrapeOptions{
			MaxPages:            paginator.MaxPages,
			Format:              p.Format,
			PaginateResults:     *p.PaginateResults,
			FetchDelay:          p.FetchDelay,
			RandomizeFetchDelay: *p.RandomizeFetchDelay,
			RetryTimes:          p.RetryTimes,
		},
	}
	return
}

//FillStruct fills s Structure with values from m map
func FillStruct(m map[string]interface{}, s interface{}) error {
	for k, v := range m {
		//	logger.Println(k,v)
		err := SetField(s, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetField(obj interface{}, name string, value interface{}) error {
	//logger.Printf("%T, %t", obj, obj)
	structValue := reflect.ValueOf(obj).Elem()
	//Value which come from json usually is in lowercase but outgoing structs may contain fields in Title Case or in UPPERCASE - f.e. URL. So we should check if there are fields in Title case or upper case before skipping non-existent fields.
	//It is unlikely there is a situation when there are several fields like url, Url, URL in the same structure.
	fValues := []reflect.Value{
		structValue.FieldByName(name),
		structValue.FieldByName(strings.Title(name)),
		structValue.FieldByName(strings.ToUpper(name)),
	}
	
	var structFieldValue reflect.Value
	for _, structFieldValue = range fValues{
		if structFieldValue.IsValid() {
			break
		}
	}
	
//	if !structFieldValue.IsValid() {
		//skip non-existent fields
//		return nil
		//return fmt.Errorf("No such field: %s in obj", name)
//	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		invalidTypeError := errors.New("Provided value type didn't match obj field type")
		return invalidTypeError
	}

	structFieldValue.Set(val)
	return nil
}