package scrape

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/spf13/viper"
)

// UnmarshalJSON casts Request interface{} type to custom Request{} type.
// If omited in Payload, Optional payload parameters initialized with default values.
// http://choly.ca/post/go-json-marshalling/
// func (p *Payload) UnmarshalJSON(data []byte) error {
// 	// type Alias Payload
// 	// aux := &struct {
// 	// 	Request fetch.Request `json:"request"`
// 	// 	*Alias
// 	// }{
// 	// 	Alias: (*Alias)(p),
// 	// }

// 	if err := json.Unmarshal(data, p); err != nil {
// 		return err
// 	}

// 	request := p.initRequest("")

// 	if p.Request.URL == "" {
// 		return &errs.BadRequest{}
// 	}
// 	// err := fillStruct(p.Request, request)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	p.Request = request

// 	//init other fields
// 	p.PayloadMD5 = utils.GenerateMD5(data)
// 	if p.Format == "" {
// 		p.Format = viper.GetString("FORMAT")
// 	}
// 	//if p.RetryTimes == 0 {
// 	//	p.RetryTimes = DefaultOptions.RetryTimes
// 	//}
// 	if p.FetchDelay == nil {
// 		delay := time.Duration(viper.GetInt("FETCH_DELAY")) * time.Millisecond
// 		p.FetchDelay = &delay
// 	}
// 	if p.RandomizeFetchDelay == nil {
// 		rand := viper.GetBool("RANDOMIZE_FETCH_DELAY")
// 		p.RandomizeFetchDelay = &rand
// 	}
// 	if p.Paginator != nil && p.Paginator.MaxPages == 0 {
// 		p.Paginator.MaxPages = viper.GetInt("MAX_PAGES")
// 	}
// 	if p.PaginateResults == nil {
// 		pag := viper.GetBool("PAGINATE_RESULTS")
// 		p.PaginateResults = &pag
// 	}
// 	return nil
// }

func (p *Payload) initRequest(newURL string) fetch.Request {
	//fetcher type from Payload structure takes precedence over FETCHER_TYPE flag value
	fetcherType := p.FetcherType
	if fetcherType == "" {
		fetcherType = viper.GetString("FETCHER_TYPE")
	}

	var URL string
	if URL = newURL; URL == "" && (fetch.Request{}) != p.Request {
		URL = p.Request.URL
	}

	//var request fetch.Request

	return fetch.Request{
		Type:           p.FetcherType,
		URL:            URL,
		FormData:       p.Request.FormData,
		UserToken:      p.Request.UserToken,
		InfiniteScroll: p.Request.InfiniteScroll,
	}



}

//fillStruct fills s Structure with values from m map
func fillStruct(m map[string]interface{}, s interface{}) error {
	for k, v := range m {
		err := setField(s, k, v)
		if err != nil {
			if k == "regexp" {
				return nil
			}
			return err
		}
	}
	//}
	return nil
}

func setField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	//Outgoing structs may contain fields in Title Case or in UPPERCASE - f.e. URL. So we should check if there are fields in Title case or upper case before skipping non-existent fields.
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
		return fmt.Errorf("Cannot set field value: %s", name)
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
