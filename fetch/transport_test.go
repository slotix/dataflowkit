package fetch

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slotix/dataflowkit/splash"
	"github.com/stretchr/testify/assert"
)

func TestDecodeSplashFetcherRequest(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"url":"http://dbconvert.com"}`)
	r := bytes.NewReader(body)
	req := httptest.NewRequest("GET", "http://example.com", r)
	actual, err := DecodeSplashFetcherRequest(ctx, req)
	assert.Nil(t, err, "Expected no error")
	var cookies []*http.Cookie
	expected := splash.Request{
		URL:      "http://dbconvert.com",
		FormData: "", Cookies: cookies, LUA: "",
	}
	assert.Equal(t, expected, actual)
}

func TestDecodeBaseFetcherRequest(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"url":"http://dbconvert.com"}`)
	r := bytes.NewReader(body)
	req := httptest.NewRequest("GET", "http://example.com", r)
	actual, err := DecodeBaseFetcherRequest(ctx, req)
	assert.Nil(t, err, "Expected no error")
	expected := BaseFetcherRequest{
		URL:    "http://dbconvert.com",
		Method: "",
	}
	assert.Equal(t, expected, actual)
}
