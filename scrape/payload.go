package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/slotix/dataflowkit/splash"
)

//http://choly.ca/post/go-json-marshalling/
//UnmarshalJSON casts Request interface{} type to custom splash.Request{} type. It initializes other payload parameters with default values.

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
	splashRequest := splash.Request{}
	err := FillStruct(aux.Request.(map[string]interface{}), &splashRequest)
	if err != nil {
		return err
	}
	p.Request = splashRequest

	//init other fields
	p.PayloadMD5 = GenerateMD5(data)
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
	return nil
}

/*
func (p Payload) PayloadToScrapeConfig() (config *Config, err error) {
	parts := []Part{}
	selectors := []string{}
	for _, f := range p.Fields {

		params := make(map[string]interface{})
		if f.Extractor.Params != nil {
			params = f.Extractor.Params.(map[string]interface{})
		}
		switch eType := f.Extractor.Type; eType {

		//For Link type Two pieces as pair Text and Attr{Attr:"href"} extractors are added.
		case "link":
			l := &extract.Link{Href: extract.Attr{Attr: "href"}}
			if params != nil {
				err := FillStruct(params, l)
				if err != nil {
					logger.Println(err)
				}
			}
			parts = append(parts, Part{
				Name:      f.Name + "_text",
				Selector:  f.Selector,
				Extractor: l.Text,
			}, Part{
				Name:      f.Name + "_link",
				Selector:  f.Selector,
				Extractor: l.Href,
			})
			//names = append(names, f.Name+"_text", f.Name+"_link")
			//Add selector just one time for link type
			selectors = append(selectors, f.Selector)

		//For image type by default Two pieces with different Attr="src" and Attr="alt" extractors will be added for field selector.
		case "image":
			i := &extract.Image{Src: extract.Attr{Attr: "src"},
				Alt: extract.Attr{Attr: "alt"}}
			if params != nil {
				err := FillStruct(params, i)
				if err != nil {
					logger.Println(err)
				}
			}
			parts = append(parts, Part{
				Name:      f.Name + "_src",
				Selector:  f.Selector,
				Extractor: i.Src,
			}, Part{
				Name:      f.Name + "_alt",
				Selector:  f.Selector,
				Extractor: i.Alt,
			})
			//	names = append(names, f.Name+"_src", f.Name+"_alt")
			//Add selector just one time for image type
			selectors = append(selectors, f.Selector)

		default:
			var e extract.Extractor
			switch eType {
			case "const":
				//	c := &extract.Const{Val: params["value"]}
				//	e = c
				e = &extract.Const{}
			case "count":
				e = &extract.Count{}
			case "text":
				e = &extract.Text{}
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
			parts = append(parts, Part{
				Name:      f.Name,
				Selector:  f.Selector,
				Extractor: e,
			})
			selectors = append(selectors, f.Selector)
			//	names = append(names, f.Name)
		}
	}

	paginator := p.Paginator
	config = &Config{
		//DividePage: scrape.DividePageBySelector(".p"),
		DividePage: DividePageByIntersection(selectors),
		Parts:      parts,
		//Paginator: paginate.BySelector(".next", "href"),
		//Header:    names,
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
*/

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
	for _, structFieldValue = range fValues {
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
