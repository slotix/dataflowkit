package fetch

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/splash"

	"golang.org/x/net/publicsuffix"
)

type UserCookies struct {
	jar *cookiejar.Jar
}

var uc UserCookies

func init() {
	uc = UserCookies{}
	jarOpts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	jar, err := cookiejar.New(jarOpts)
	if err != nil {
		logger.Error("Failed to create Cookie Jar")

	} else {
		uc.jar = jar
	}
}

//Type represents types of fetcher
type Type string

//Fetcher types
const (
	//Base fetcher is used for downloading html web page using Go standard library's http
	Base Type = "Base"
	//Splash server is used to download content of web page after running of js scripts on the web page.
	Splash = "Splash"
)

// Fetcher is the interface that must be satisfied by things that can fetch
// remote URLs and return their contents.
//
// Note: Fetchers may or may not be safe to use concurrently.  Please read the
// documentation for each fetcher for more details.
type Fetcher interface {
	// Prepare is called once at the beginning of the scrape.
	Prepare() error
	//Response return response after fetch Request.
	Response(request FetchRequester) (FetchResponser, error)
	// Fetch is called to retrieve HTML content of a document from the remote server.
	Fetch(request FetchRequester) (io.ReadCloser, error)
	// Close is called when the scrape is finished, and can be used to clean up
	// allocated resources or perform other cleanup actions.
	Close()
}

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
	//GetStatusCode returns status code after document fetch
	GetStatusCode() int
	//GetHTML return html content of fetched document
	GetHTML() (io.ReadCloser, error)
	//GetHeaders returns Headers from response
	GetHeaders() http.Header
}

//FetchRequester interface interface that must be satisfied the listed methods
type FetchRequester interface {
	//GetURL returns initial URL from Request
	GetURL() string
	// Host returns Host value from Request
	Host() (string, error)
	SetCookies(string)
	//SetURL initializes URL value of Request
	SetURL(string)
	//Returns Params (FormData)
	GetParams() string
	//Type
	Type() string
}

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
func (sf *SplashFetcher) Fetch(request FetchRequester) (io.ReadCloser, error) {
	r, err := sf.Response(request)
	if err != nil {
		return nil, err
	}
	return r.GetHTML()
}

//Response return response from Splash server after document fetching
func (sf *SplashFetcher) Response(request FetchRequester) (FetchResponser, error) {
	req := request.(splash.Request)
	u, err := url.Parse(req.GetURL())
	if err != nil {
		return nil, err
	}
	req.Cookies = uc.jar.Cookies(u)
	r, err := req.GetResponse()
	if err != nil {
		return nil, err
	}

	cArr := []*http.Cookie{}
	for _, c := range r.Cookies {
		cookie := c.Cookie
		cArr = append(cArr, &cookie)
	}
	uc.jar.SetCookies(u, cArr)
	return r, nil
}

// Close is called when the scrape is finished, and can be used to clean up
// allocated resources or perform other cleanup actions.
func (sf *SplashFetcher) Close() {
	return
}

// Static type assertion
var _ Fetcher = &SplashFetcher{}

// NewBaseFetcher creates instances of NewBaseFetcher{} to fetch
// a page content from regular websites as-is
// without running js scripts on the page.
func NewBaseFetcher() (*BaseFetcher, error) {
	client := &http.Client{Jar: uc.jar}

	bf := &BaseFetcher{
		client: client,
	}
	return bf, nil
}

// Prepare is called once at the beginning of the scrape.
func (bf *BaseFetcher) Prepare() error {
	if bf.PrepareClient != nil {
		return bf.PrepareClient(bf.client)
	}
	return nil
}

//Fetch retrieves document from the remote server. It returns web page content along with cache and expiration information.
func (bf *BaseFetcher) Fetch(request FetchRequester) (io.ReadCloser, error) {
	r, err := bf.Response(request)
	if err != nil {
		return nil, err
	}
	return r.GetHTML()
}

func parseParams(params string) url.Values {
	//"auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=usr&ips_password=passw&rememberMe=0"
	formData := url.Values{}
	pairs := strings.Split(params, "&")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		formData.Add(kv[0], kv[1])
	}
	return formData
}

//Response return response after document fetching using BaseFetcher
func (bf *BaseFetcher) Response(request FetchRequester) (FetchResponser, error) {
	//URL validation
	if _, err := url.ParseRequestURI(strings.TrimSpace(request.GetURL())); err != nil {
		return nil, &errs.BadRequest{err}
	}

	r := request.(BaseFetcherRequest)
	var err error
	var req *http.Request
	var resp *http.Response
	if request.GetParams() == "" {
		req, err = http.NewRequest(r.Method, r.URL, nil)
		if err != nil {
			return nil, err
		}
	} else {
		r.Method = "POST"
		formData := parseParams(r.GetParams())
		req, err = http.NewRequest(r.Method, r.URL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	}
	if bf.PrepareRequest != nil {
		if err = bf.PrepareRequest(req); err != nil {
			return nil, err
		}
	}

	resp, err = bf.client.Do(req)
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
		case 500:
			return nil, &errs.InternalServerError{}
		case 504:
			return nil, &errs.GatewayTimeout{}
		case 401:
			return nil, &errs.Unauthorized{}
		default:
			return nil, &errs.Error{"Unknown Error"}
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := BaseFetcherResponse{
		Response:   resp,
		URL:        resp.Request.URL.String(),
		HTML:       string(body),
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
	// gCurCookies = bf.client.Jar.Cookies(req.URL)
	// cookieNum = len(gCurCookies)
	// log.Printf("cookieNum=%d", cookieNum)
	// for i := 0; i < cookieNum; i++ {
	// 	var curCk *http.Cookie = gCurCookies[i]
	// 	//log.Printf("curCk.Raw=%s", curCk.Raw)
	// 	log.Printf("Cookie [%d]", i)
	// 	log.Printf("Name\t=%s", curCk.Name)
	// 	log.Printf("Value\t=%s", curCk.Value)
	// 	log.Printf("Path\t=%s", curCk.Path)
	// 	log.Printf("Domain\t=%s", curCk.Domain)
	// 	log.Printf("Expires\t=%s", curCk.Expires)
	// 	log.Printf("RawExpires=%s", curCk.RawExpires)
	// 	log.Printf("MaxAge\t=%d", curCk.MaxAge)
	// 	log.Printf("Secure\t=%t", curCk.Secure)
	// 	log.Printf("HttpOnly=%t", curCk.HttpOnly)
	// 	log.Printf("Raw\t=%s", curCk.Raw)
	// 	log.Printf("Unparsed=%s", curCk.Unparsed)
	// }
	return &response, err
}

// Close is called when the scrape is finished, and can be used to clean up
// allocated resources or perform other cleanup actions.
func (bf *BaseFetcher) Close() {
	return
}

// Static type assertion
var _ Fetcher = &BaseFetcher{}
