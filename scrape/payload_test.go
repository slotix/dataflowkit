package scrape

import (
	"testing"
	"time"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/utils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var data []byte

func init() {
	data = []byte(`
		{
			"name":"collection",
			"request":{
			   "url":"https://example.com"
			},
			"fields":[
			   {
				  "name":"Title",
				  "selector":".product-container a",
				  "extractor":{
					 "types":["text", "href"],
					 "filters":[
						"trim",
						"lowerCase"
					 ],
					 "params":{
						"includeIfEmpty":false
					 }
				  }
			   },
			   {
				  "name":"Image",
				  "selector":"#product-container img",
				  "extractor":{
					 "types":["alt","src","width","height"],
					 "filters":[
						"trim",
						"upperCase"
					 ]
				  }
			   },
			   {
				  "name":"Buyinfo",
				  "selector":".buy-info",
				  "extractor":{
					 "types":["text"],
					 "params":{
						"includeIfEmpty":false
					 }
				  }
			   }
			],
			"paginator":{
			   "selector":".next",
			   "attr":"href",
			   "maxPages":0
			}
		   }
		`)
}

func TestPayload_UnmarshalJSON_Req_nil(t *testing.T) {
	for _, fType := range []string{
		"splash",
		"base",
	} {

		viper.Set("FETCHER_TYPE", fType)
		p := &Payload{}
		err := p.UnmarshalJSON(data)
		assert.NoError(t, err)
		assert.Equal(t, p.Name, "collection")
		switch fType {
		case "splash":
			assert.Equal(t, p.Request, &splash.Request{URL: "https://example.com"})
		case "base":
			assert.Equal(t, p.Request, &fetch.BaseFetcherRequest{URL: "https://example.com"})
		}
		assert.Equal(t, p.Fields,
			[]Field{
				Field{
					Name:     "Title",
					Selector: ".product-container a",
					Extractor: Extractor{
						Types:  []string{"text", "href"},
						Params: map[string]interface{}{"includeIfEmpty": false}, Filters: []string{"trim", "lowerCase"}},
					Details: (*details)(nil)},
				Field{
					Name:     "Image",
					Selector: "#product-container img",
					Extractor: Extractor{
						Types:   []string{"alt", "src", "width", "height"},
						Params:  map[string]interface{}(nil),
						Filters: []string{"trim", "upperCase"}},
					Details: (*details)(nil)},
				Field{
					Name:     "Buyinfo",
					Selector: ".buy-info",
					Extractor: Extractor{
						Types:  []string{"text"},
						Params: map[string]interface{}{"includeIfEmpty": false}, Filters: []string(nil)},
					Details: (*details)(nil)}},
		)
		assert.Equal(t, p.Paginator,
			&paginator{
				Selector:  ".next",
				Attribute: "href",
				MaxPages:  0,
			})
		//assert.Equal(t, p.Format, "json")
		pr := false
		
		assert.Equal(t, p.PaginateResults, &pr)
		assert.Equal(t, p.PayloadMD5, utils.GenerateMD5(data))
		td := time.Duration(0)
		assert.Equal(t, p.FetchDelay, &td)
		rfd := false
		assert.Equal(t, p.PaginateResults, &rfd)
	}

}

func TestPayload_UnmarshalJSON_Req_not_nil(t *testing.T) {
	for _, p := range []Payload{
		Payload{
			Request:     &fetch.BaseFetcherRequest{URL: "https://example.com"},
			FetcherType: "base",
		},
		Payload{
			Request: &splash.Request{
				URL: "https://example.com"},
			FetcherType: "splash",
			Paginator: &paginator{
				InfiniteScroll: true,
			},
		},
	} {

		err := p.UnmarshalJSON(data)
		assert.NoError(t, err)
		//assert.Equal(t, p.Name, "collection")
	}
}

func TestPayload_UnmarshalJSON_Invalid_Request(t *testing.T) {
	p := &Payload{}
	err := p.UnmarshalJSON([]byte{})
	assert.Error(t, err)

	//invalid fetcher type
	viper.Set("FETCHER_TYPE", "")
	p = &Payload{}
	err = p.UnmarshalJSON([]byte(`
		{
			"name":"Bad collection",
			"request":{
				"url":"https://example.com"
			 }
		}
	`))
	assert.Error(t, err)
	t.Log(err)
}
