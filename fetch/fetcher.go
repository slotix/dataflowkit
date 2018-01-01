package fetch

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/splash"

	"golang.org/x/net/publicsuffix"
)

//Type represents types of fetcher
type Type string

//Fetcher types
const (
	//Base fetcher is used for downloading html web page using Go standard library's http
	Base Type = "Base"
	//Splash server is used to download content of web page after running of js scripts on the web page.
	Splash = "Splash"
)

//NewFetcher creates instances of Fetcher for downloading a web page.
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

// SplashFetcher is a Fetcher that uses Scrapinghub splash
// to fetch URLs. Splash is a javascript rendering service
// Read more at https://github.com/scrapinghub/splash
type SplashFetcher struct {
	//client *http.Client

	// PrepareSplash is called once at the beginning of the scrape. It is not used currently
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

// NewSplashFetcher creates instances of SplashFetcher{} to fetch a page content from remote Scrapinghub splash service.
func NewSplashFetcher() (*SplashFetcher, error) {
	// Set up the HTTP client
	// jarOpts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	// jar, err := cookiejar.New(jarOpts)
	// if err != nil {
	// 	return nil, err
	// }
	// client := &http.Client{Jar: jar}

	// ret := &SplashFetcher{
	// 	client: client,
	// }
	// return ret, nil
	sf := &SplashFetcher{}
	return sf, nil
}

// Prepare is called once at the beginning of the scrape.
func (sf *SplashFetcher) Prepare() error {
	if sf.PrepareSplash != nil {
		return sf.PrepareSplash()
	}
	return nil
}

//Fetch retrieves document from the remote server. It returns web page content along with cache and expiration information.
func (sf *SplashFetcher) Fetch(request FetchRequester) (FetchResponser, error) {
	req := request.(splash.Request)
	r, err := req.GetResponse()
	if err != nil {
		return nil, err
	}
	return r, nil

}

// Static type assertion
var _ Fetcher = &SplashFetcher{}

// Close is called when the scrape is finished, and can be used to clean up
// allocated resources or perform other cleanup actions.
func (sf *SplashFetcher) Close() {
	return
}

// NewBaseFetcher creates instances of NewBaseFetcher{} to fetch
// a page content from regular websites as-is
// without running js scripts on the page.
// f.e. robots.txt are retrieved with BaseFetcher
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

// Prepare is called once at the beginning of the scrape.
func (bf *BaseFetcher) Prepare() error {
	if bf.PrepareClient != nil {
		return bf.PrepareClient(bf.client)
	}
	return nil
}

//Fetch retrieves document from the remote server. It returns web page content along with cache and expiration information.
func (bf *BaseFetcher) Fetch(request FetchRequester) (FetchResponser, error) {
	err := request.Validate()
	if err != nil {
		return nil, err
	}
	r := request.(BaseFetcherRequest)
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
		return nil, &errs.BadRequest{err}
	}

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 404:
			return nil, &errs.NotFound{r.URL}
		case 403:
			return nil, &errs.Forbidden{r.URL}
		case 400:
			return nil, &errs.BadRequest{err}
		default:
			return nil, err
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := BaseFetcherResponse{
		Response:   resp,
		URL:        resp.Request.URL.String(),
		HTML:       body,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}

	//set Cache control parameters
	response.SetCacheInfo()

	if bf.ProcessResponse != nil {
		if err = bf.ProcessResponse(resp); err != nil {
			return nil, err
		}
	}
	return &response, nil
}

// Close is called when the scrape is finished, and can be used to clean up
// allocated resources or perform other cleanup actions.
func (bf *BaseFetcher) Close() {
	return
}

// Static type assertion
var _ Fetcher = &BaseFetcher{}

//FetchResponser interface that must be satisfied the listed methods
type FetchResponser interface {
	//Returns expires value of response
	GetExpires() time.Time
	//Returns an array of reasons why a response should not be cached if any.
	GetReasonsNotToCache() []cacheobject.Reason
	//ReasonsNotToCache and Expires values are set here
	SetCacheInfo()
	//GetURL returns final URL after all redirects
	GetURL() string
}

//FetchRequester interface interface that must be satisfied the listed methods
type FetchRequester interface {
	//GetURL returns initial URL from Request
	GetURL() string
	//Validates request before sending
	Validate() error
	URL2MD5() string
}
