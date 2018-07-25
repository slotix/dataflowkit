package parse

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// var payload1 = scrape.Payload{
// 	Name: "test",
// 	Request: fetch.Request{
// 		Type:      "base",
// 		URL:       "http://books.toscrape.com",
// 		UserToken: "12345",
// 	},
// 	Fields: []scrape.Field{
// 		scrape.Field{
// 			Name:     "Title",
// 			Selector: "h3 a",
// 			Extractor: scrape.Extractor{
// 				Types:   []string{"text", "href"},
// 				Filters: []string{"trim"},
// 			},
// 		},
// 		scrape.Field{
// 			Name:     "Price",
// 			Selector: ".price_color",
// 			Extractor: scrape.Extractor{
// 				Types: []string{"regex"},
// 				Params: map[string]interface{}{
// 					"regexp": "([\\d\\.]+)",
// 				},
// 				Filters: []string{"trim"},
// 			},
// 		},
// 		scrape.Field{
// 			Name:     "Image",
// 			Selector: ".thumbnail",
// 			Extractor: scrape.Extractor{
// 				Types:   []string{"src", "alt"},
// 				Filters: []string{"trim"},
// 			},
// 		},
// 	},
// 	Format: "json",
// }

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	//healthCheckHandler(w, req)
	handler := http.HandlerFunc(HealthCheckHandler)
	handler.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, []byte(`{"alive": true}`), body)
}
