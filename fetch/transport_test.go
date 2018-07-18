package fetch

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	//healthCheckHandler(w, req)
	handler := http.HandlerFunc(healthCheckHandler)
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

func TestQuery(t *testing.T) {
	url := "http://localhost/test?q=http%3A%2F%2Fgoogle.com"

	hit := false
	m := mux.NewRouter()
	m.Path("/test").Queries("q", "{q}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
	})

	req, _ := http.NewRequest("GET", url, nil)
	m.ServeHTTP(&httptest.ResponseRecorder{}, req)
	if !hit {
		t.Errorf("query did not hit")
	}
}
