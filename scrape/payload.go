package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/utils"
	"github.com/spf13/viper"
)

// UnmarshalJSON casts Request interface{} type to custom splash.Request{} type.
// If omited in Payload, Optional payload parameters initialized with default values.
// http://choly.ca/post/go-json-marshalling/
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

	request, err := p.initRequest("")
	if err != nil {
		return err
	}
	if aux.Request == nil {
		return &errs.BadRequest{}
	}
	err = fillStruct(aux.Request.(map[string]interface{}), request)
	if err != nil {
		return err
	}
	p.Request = request

	//init other fields
	p.PayloadMD5 = utils.GenerateMD5(data)
	if p.Format == "" {
		p.Format = viper.GetString("FORMAT")
	}
	//if p.RetryTimes == 0 {
	//	p.RetryTimes = DefaultOptions.RetryTimes
	//}
	if p.FetchDelay == nil {
		delay := time.Duration(viper.GetInt("FETCH_DELAY")) * time.Millisecond
		p.FetchDelay = &delay
	}
	if p.RandomizeFetchDelay == nil {
		rand := viper.GetBool("RANDOMIZE_FETCH_DELAY")
		p.RandomizeFetchDelay = &rand
	}
	if p.Paginator != nil && p.Paginator.MaxPages == 0 {
		p.Paginator.MaxPages = viper.GetInt("MAX_PAGES")
	}
	if p.PaginateResults == nil {
		pag := viper.GetBool("PAGINATE_RESULTS")
		p.PaginateResults = &pag
	}
	return nil
}

func (p *Payload) initRequest(newURL string) (fetch.FetchRequester, error) {
	//fetcher type from Payload structure takes precedence over FETCHER_TYPE flag value
	fetcherType := p.FetcherType
	if fetcherType == "" {
		fetcherType = viper.GetString("FETCHER_TYPE")
	}

	var URL string
	if URL = newURL; URL == "" && p.Request != nil {
		URL = p.Request.GetURL()
	}

	var request fetch.FetchRequester
	switch strings.ToLower(fetcherType) {
	case "splash":
		if p.Request == nil {
			request = &splash.Request{}
		} else {
			var infiniteScroll bool
			if infiniteScroll = false; p.Paginator != nil && p.Paginator.InfiniteScroll {
				infiniteScroll = true
			}
			request = &splash.Request{
				URL:            URL,
				FormData:       p.Request.GetFormData(),
				UserToken:      p.Request.GetUserToken(),
				InfiniteScroll: infiniteScroll}
		}
	case "base":
		if p.Request == nil {
			request = &fetch.BaseFetcherRequest{}
		} else {
			request = &fetch.BaseFetcherRequest{
				URL:       URL,
				FormData:  p.Request.GetFormData(),
				UserToken: p.Request.GetUserToken()}
		}
	default:
		err := errors.New("invalid fetcher type specified")
		logger.Error(err.Error())
		return nil, err
	}
	return request, nil
}

//fillStruct fills s Structure with values from m map
func fillStruct(m map[string]interface{}, s interface{}) error {
	for k, v := range m {
		err := setField(s, k, v)
		if err != nil {
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
