package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/slotix/dataflowkit/crypto"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)

// UnmarshalJSON casts Request interface{} type to custom splash.Request{} type.
// If optional payload parameters are not passed along with obligatorily ones
// they are initialized with default values.
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
	splashRequest := splash.Request{}
	err := fillStruct(aux.Request.(map[string]interface{}), &splashRequest)
	if err != nil {
		return err
	}
	p.Request = splashRequest

	//init other fields
	p.PayloadMD5 = crypto.GenerateMD5(data)
	if p.Format == "" {
		p.Format = viper.GetString("FORMAT")
	}
	//if p.RetryTimes == 0 {
	//	p.RetryTimes = DefaultOptions.RetryTimes
	//}
	if p.FetchDelay == 0 {
		p.FetchDelay = time.Duration(viper.GetInt("FETCH_DELAY")) * time.Millisecond
	}
	if p.RandomizeFetchDelay == nil {
		rand := viper.GetBool("RANDOMIZE_FETCH_DELAY")
		p.RandomizeFetchDelay = &rand
	}
	if p.Paginator.MaxPages == 0{
		p.Paginator.MaxPages = viper.GetInt("MAX_PAGES")
	}
	if p.PaginateResults == nil {
		pag := viper.GetBool("PAGINATE_RESULTS")
		p.PaginateResults = &pag
	}
	return nil
}

//fillStruct fills s Structure with values from m map
func fillStruct(m map[string]interface{}, s interface{}) error {
	for k, v := range m {
		//switch v.(type) {
		//	case []interface{}:

		//determine the type of array

		//		value := []string{}
		//		for _, item := range v.([]interface{}) {
		//				logger.Infof("%T", item)
		//TODO: casting to string should be exchanged with something more universal.
		//			value = append(value, item.(string))
		//		}
		//
		//		err := setField(s, k, value)
		//		if err != nil {
		//			return err
		//			}
		//		default:
		err := setField(s, k, v)
		if err != nil {
			return err
		}
	}
	//}
	return nil
}

func setField(obj interface{}, name string, value interface{}) error {
	//logger.Info("%T, %t", obj, obj)
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
