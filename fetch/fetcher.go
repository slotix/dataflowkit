package fetch

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/slotix/dataflowkit/splash"

	"golang.org/x/net/publicsuffix"
)

type Type string

const (
	Base   Type = "Base"
	Splash      = "Splash"
)

func NewFetcher(t Type) (fetcher Fetcher, err error) {
	switch t {
	case Base:
		fetcher, err = NewBaseFetcher()
		return
	case Splash:
		fetcher, err = NewSplashFetcher()
		return
	default:
		return nil, errors.New("Can't create Fetcher")
	}
}

// Fetcher is the interface that must be satisfied by things that can fetch
// remote URLs and return their contents.
//
// Note: Fetchers may or may not be safe to use concurrently.  Please read the
// documentation for each fetcher for more details.
type Fetcher interface {
	// Prepare is called once at the beginning of the scrape.
	Prepare() error

	// Fetch is called to retrieve a document from the remote server.
	Fetch(request FetchRequester) (FetchResponser, error)

	// Close is called when the scrape is finished, and can be used to clean up
	// allocated resources or perform other cleanup actions.
	Close()
}

// BaseFetcher is a Fetcher that uses the Go standard library's http
// client to fetch URLs.
type BaseFetcher struct {
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


//Fetch retrieves document from the remote server. It returns splash.Response as it is not enough to get just page content but during scraping sessions auxiliary information like cookies should be avaialable.
func (sf *SplashFetcher) Fetch(request FetchRequester) (FetchResponser, error) {
	req := request.(splash.Request)
	r, err := req.GetResponse()
	if err != nil {
		return nil, err
	}
	return r, nil

}

var _ Fetcher = &SplashFetcher{}

func (sf *SplashFetcher) Close() {
	return
}

func NewBaseFetcher() (*BaseFetcher, error) {
	// Set up the HTTP client
	jarOpts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	jar, err := cookiejar.New(jarOpts)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}

	ret := &BaseFetcher{
		client: client,
	}
	return ret, nil
}

func (bf *BaseFetcher) Prepare() error {
	if bf.PrepareClient != nil {
		return bf.PrepareClient(bf.client)
	}
	return nil
}

func (bf *BaseFetcher) Fetch(request FetchRequester) (FetchResponser, error) {	
	err := request.Validate()
	if err != nil {
		return nil, err
	}
	r := request.(BaseFetcherRequest)
	r.Method = "POST"
	req, err := http.NewRequest(r.Method, r.URL, nil)
	if err != nil {
		return nil, err
	}

	if bf.PrepareRequest != nil {
		if err = bf.PrepareRequest(req); err != nil {
			return nil, err
		}
	}

	resp, err := bf.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := BaseFetcherResponse{Response: resp, HTML: body, StatusCode: resp.StatusCode, Status: resp.Status}

	//is resource cacheable ?
	response.SetCacheInfo()

	if bf.ProcessResponse != nil {
		if err = bf.ProcessResponse(resp); err != nil {
			return nil, err
		}
	}

	//return resp.Body.(io.ReadCloser), nil
	//return resp.Body, nil
	return &response, nil
}

func (bf *BaseFetcher) Close() {
	return
}

// Static type assertion
var _ Fetcher = &BaseFetcher{}

type FetchResponser interface {
	GetExpires() time.Time
	GetCacheable() bool
	SetCacheInfo()
}

type FetchRequester interface {
	GetURL() string
	Validate() error
}
