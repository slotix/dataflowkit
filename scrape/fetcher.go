package scrape

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/splash"

	"golang.org/x/net/publicsuffix"
)

// Fetcher is the interface that must be satisfied by things that can fetch
// remote URLs and return their contents.
//
// Note: Fetchers may or may not be safe to use concurrently.  Please read the
// documentation for each fetcher for more details.
type Fetcher interface {
	// Prepare is called once at the beginning of the scrape.
	Prepare() error

	// Fetch is called to retrieve each document from the remote server.
	//Fetch(method, url string) (io.ReadCloser, error)
	Fetch(request interface{}) (interface{}, error)

	// Close is called when the scrape is finished, and can be used to clean up
	// allocated resources or perform other cleanup actions.
	Close()
}

// HttpClientFetcher is a Fetcher that uses the Go standard library's http
// client to fetch URLs.
type HttpClientFetcher struct {
	client *http.Client

	// PrepareClient prepares this fetcher's http.Client for usage.  Use this
	// function to do things like logging in.  If the function returns an error,
	// the scrape is aborted.
	PrepareClient func(*http.Client) error

	// PrepareRequest prepares each request that will be sent, prior to sending.
	// This is useful for, e.g. setting custom HTTP headers, changing the User-
	// Agent, and so on.  If the function returns an error, then the scrape will
	// be aborted.
	//
	// Note: this function does NOT apply to requests made during the
	// PrepareClient function (above).
	PrepareRequest func(*http.Request) error

	// ProcessResponse modifies a response that is returned from the server before
	// it is handled by the scraper.  If the function returns an error, then the
	// scrape will be aborted.
	ProcessResponse func(*http.Response) error
}

// SplashClientFetcher is a Fetcher that uses Scrapinghub splash
// to fetch URLs. Splash is a javascript rendering service
type SplashFetcher struct {
//	request *splash.Request

	PrepareSplash func() error
	// PrepareRequest prepares each request that will be sent, prior to sending.
	// This is useful for, e.g. setting custom HTTP headers, changing the User-
	// Agent, and so on.  If the function returns an error, then the scrape will
	// be aborted.
	//
	// Note: this function does NOT apply to requests made during the
	// PrepareClient function (above).
	PrepareRequest func(*splash.Request) error
}

func NewSplashFetcher() (*SplashFetcher, error) {
	sf := &SplashFetcher{}
	return sf, nil
}

func (sf *SplashFetcher) Prepare() error {
	if sf.PrepareSplash != nil {
		return sf.PrepareSplash()
	}
	return nil
}

//ValidateRequest validates each request that will be sent, prior to sending.
func (sf *SplashFetcher) ValidateRequest(req *splash.Request) error {
	//req.URL normalization and validation
	//	request := req.(splash.Request)
	reqURL := strings.TrimSpace(req.URL)
	if _, err := url.ParseRequestURI(reqURL); err != nil {
		return &errs.BadRequest{err}
	}
	req.URL = reqURL
	return nil
}

//Fetch retrieves document from the remote server. It returns splash.Response as it is not enough to get just page content but during scraping sessions auxiliary information like cookies should be avaialable.
func (sf *SplashFetcher) Fetch(request interface{}) (interface{}, error) {

	//r, err := splash.GetResponse(request.(splash.Request))
	r, err := splash.GetResponse(request.(splash.Request))

	if err != nil {
		return nil, err
	}

	return r, nil

}

var _ Fetcher = &SplashFetcher{}

func (sf *SplashFetcher) Close() {
	return
}

func NewHttpClientFetcher() (*HttpClientFetcher, error) {
	// Set up the HTTP client
	jarOpts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	jar, err := cookiejar.New(jarOpts)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}

	ret := &HttpClientFetcher{
		client: client,
	}
	return ret, nil
}

func (hf *HttpClientFetcher) Prepare() error {
	if hf.PrepareClient != nil {
		return hf.PrepareClient(hf.client)
	}
	return nil
}

type HttpClientFetcherRequest struct {
	URL    string
	Method string
}

func (hf *HttpClientFetcher) Fetch(request interface{}) (interface{}, error) {
	//	var r HttpClientFetcherRequest
	//	err := json.Unmarshal(request, &r)
	//	if err != nil {
	//		return nil, err
	//	}
	r := request.(HttpClientFetcherRequest)
	req, err := http.NewRequest(r.Method, r.URL, nil)
	if err != nil {
		return nil, err
	}

	if hf.PrepareRequest != nil {
		if err = hf.PrepareRequest(req); err != nil {
			return nil, err
		}
	}

	resp, err := hf.client.Do(req)
	if err != nil {
		return nil, err
	}

	if hf.ProcessResponse != nil {
		if err = hf.ProcessResponse(resp); err != nil {
			return nil, err
		}
	}

	//return resp.Body.(io.ReadCloser), nil
	//return resp.Body, nil
	return resp, nil
}

func (hf *HttpClientFetcher) Close() {
	return
}

/*
func (hf *HttpClientFetcher) GetURL(request interface{}) string {
	logger.Println(request)
	return request.(HttpClientFetcherRequest).URL
}
*/
// Static type assertion
var _ Fetcher = &HttpClientFetcher{}
