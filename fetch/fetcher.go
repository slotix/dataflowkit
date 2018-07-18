package fetch

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/juju/persistent-cookiejar"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/rpcc"
	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)

//Type represents types of fetcher
type Type string

//Fetcher types
const (
	//Base fetcher is used for downloading html web page using Go standard library's http
	Base Type = "Base"
	//Splash server is used to download content of web page after running of js scripts on the web page.
	Splash = "Splash"
	//Headless chrome is used to download content from JS driven web pages
	Chrome = "Chrome"
)

// Fetcher is the interface that must be satisfied by things that can fetch
// remote URLs and return their contents.
//
// Note: Fetchers may or may not be safe to use concurrently.  Please read the
// documentation for each fetcher for more details.
type Fetcher interface {
	//  Response return response after fetch Request.
	//Response(request FetchRequester) (FetchResponser, error)
	//  Fetch is called to retrieve HTML content of a document from the remote server.
	Fetch(request FetchRequester) (io.ReadCloser, error)
	GetCookieJar() *cookiejar.Jar
	SetCookieJar(jar *cookiejar.Jar)
}

//FetchResponser interface that must be satisfied the listed methods
type FetchResponser interface {
	//  GetExpires returns expires value of response
	GetExpires() time.Time
	//  GetReasonsNotToCache returns an array of reasons why a response should not be cached if any.
	GetReasonsNotToCache() []cacheobject.Reason
	//  ReasonsNotToCache and Expires values are set here
	SetCacheInfo()
	//  GetURL returns final URL after all redirects
	GetURL() string
	//  GetHTML return html content of fetched document
	GetHTML() (io.ReadCloser, error)
}

//FetchRequester interface interface that must be satisfied the listed methods
type FetchRequester interface {
	//  GetURL returns initial URL from Request
	GetURL() string
	//  Host returns Host value from Request
	Host() (string, error)
	//  GetFormData Returns Form Data from FetchRequester
	GetFormData() string
	//  Type return type of request : base or splash
	Type() string
	GetUserToken() string
}

//NewFetcher creates instances of Fetcher for downloading a web page.
func NewFetcher(t Type) Fetcher {
	switch t {
	case Base:
		return NewBaseFetcher()
	case Splash:
		return NewSplashFetcher()
	case Chrome:
		return NewChromeFetcher()
	default:
		logger.Panicf("unhandled type: %#v", t)
	}
	panic("unreachable")
}

// BaseFetcher is a Fetcher that uses the Go standard library's http
// client to fetch URLs.
type BaseFetcher struct {
	client *http.Client
	jar    *cookiejar.Jar
}

// SplashFetcher is a Fetcher that uses Scrapinghub splash
// to fetch URLs. Splash is a javascript rendering service
// Read more at https://github.com/scrapinghub/splash
type SplashFetcher struct {
	jar *cookiejar.Jar
}

// ChromeFetcher is used to fetch Java Script rendeded pages.
type ChromeFetcher struct {
	cdpClient *cdp.Client
	client    *http.Client
	jar       *cookiejar.Jar
}

// NewSplashFetcher creates instance of SplashFetcher{} to fetch a page content from remote Scrapinghub splash service.
func NewSplashFetcher() *SplashFetcher {
	sf := &SplashFetcher{}
	return sf
}

// Fetch retrieves document from the remote server. It returns web page content along with cache and expiration information.
func (sf *SplashFetcher) Fetch(request FetchRequester) (io.ReadCloser, error) {
	r, err := sf.Response(request)
	if err != nil {
		return nil, err
	}
	return r.GetHTML()
}

// Response return response from Splash server after document fetching
func (sf *SplashFetcher) Response(request FetchRequester) (FetchResponser, error) {
	req := request.(splash.Request)
	u, err := url.Parse(req.GetURL())
	if err != nil {
		return nil, err
	}

	if sf.jar != nil {
		req.Cookies = sf.jar.AllCookies()
	}

	r, err := req.GetResponse()
	if err != nil {
		return nil, err
	}

	cArr := []*http.Cookie{}
	for _, c := range r.Cookies {
		cookie := c.Cookie
		cArr = append(cArr, &cookie)
	}
	sf.jar.SetCookies(u, cArr)
	return r, nil
}

func (sf *SplashFetcher) GetCookieJar() *cookiejar.Jar {
	return sf.jar
}

func (sf *SplashFetcher) SetCookieJar(jar *cookiejar.Jar) {
	sf.jar = jar
}

// Static type assertion
var _ Fetcher = &SplashFetcher{}

// NewChromeFetcher returns ChromeFetcher
func NewChromeFetcher() *ChromeFetcher {
	var client *http.Client
	proxy := viper.GetString("PROXY")
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			logger.Error(err)
			return nil
		}
		transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}
	f := &ChromeFetcher{
		client: client,
	}
	return f
}

// Fetch retrieves document from the remote server. It returns web page content along with cache and expiration information.
func (f *ChromeFetcher) Fetch(request FetchRequester) (io.ReadCloser, error) {
	//URL validation
	if _, err := url.ParseRequestURI(strings.TrimSpace(request.GetURL())); err != nil {
		return nil, &errs.BadRequest{err}
	}

	if f.jar != nil {
		f.client.Jar = f.jar
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	devt := devtool.New("http://localhost:9222", devtool.WithClient(f.client))
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		return nil, err
	}
	// Connect to WebSocket URL (page) that speaks the Chrome Debugging Protocol.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer conn.Close() // Cleanup.
	// Create a new CDP Client that uses conn.
	f.cdpClient = cdp.NewClient(conn)

	// Give enough capacity to avoid blocking any event listeners
	abort := make(chan error, 2)
	// Watch the abort channel.
	go func() {
		select {
		case <-ctx.Done():
		case err := <-abort:
			fmt.Printf("aborted: %s\n", err.Error())
			cancel()
		}
	}()
	// Setup event handlers early because domain events can be sent as
	// soon as Enable is called on the domain.
	// if err = abortOnErrors(ctx, c, scriptID, abort); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	if err = runBatch(
		// Enable all the domain events that we're interested in.
		func() error { return f.cdpClient.DOM.Enable(ctx) },
		func() error { return f.cdpClient.Network.Enable(ctx, nil) },
		func() error { return f.cdpClient.Page.Enable(ctx) },
		func() error { return f.cdpClient.Runtime.Enable(ctx) },

		//func() error { return setCookies(ctx, c.Network, Cookies...) },
	); err != nil {
		return nil, err
	}
	domLoadTimeout := 5 * time.Second
	if request.GetFormData() == "" {
		err = f.navigate(ctx, f.cdpClient.Page, "GET", request.GetURL(), "", domLoadTimeout)
		if err != nil {
			return nil, err
		}
	} else {
		formData := parseFormData(request.GetFormData())
		err = f.navigate(ctx, f.cdpClient.Page, "POST", request.GetURL(), formData.Encode(), domLoadTimeout)
	}

	//TODO: add main loader script
	// err = f.runJSFromFile(ctx, "./chrome/loader.js")
	// if err != nil {
	// 	return nil, err
	// }

	//TODO: scroll2bottom
	// if request.InfiniteScroll {
	// 	err = f.runJSFromFile(ctx, "./chrome/scroll2bottom.js")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	doc, err := f.cdpClient.DOM.GetDocument(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Get the outer HTML for the page.
	result, err := f.cdpClient.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
		NodeID: &doc.Root.NodeID,
	})
	if err != nil {
		return nil, err
	}
	readCloser := ioutil.NopCloser(strings.NewReader(result.OuterHTML))
	return readCloser, nil

}

func (cf *ChromeFetcher) SetCookieJar(jar *cookiejar.Jar) {
	cf.jar = jar
}

func (cf *ChromeFetcher) GetCookieJar() *cookiejar.Jar {
	return cf.jar
}

// Static type assertion
var _ Fetcher = &ChromeFetcher{}

// NewBaseFetcher creates instances of NewBaseFetcher{} to fetch
// a page content from regular websites as-is
// without running js scripts on the page.
func NewBaseFetcher() *BaseFetcher {
	var client *http.Client
	proxy := viper.GetString("PROXY")
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			logger.Error(err)
			return nil
		}
		transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}
	f := &BaseFetcher{
		client: client,
	}
	return f
}

// Fetch retrieves document from the remote server. It returns web page content along with cache and expiration information.
func (bf *BaseFetcher) Fetch(request FetchRequester) (io.ReadCloser, error) {
	r, err := bf.Response(request)
	if err != nil {
		return nil, err
	}
	return r.GetHTML()
}

// parseFormData is used for converting formdata string to url.Values type
func parseFormData(fd string) url.Values {
	//"auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=usr&ips_password=passw&rememberMe=0"
	formData := url.Values{}
	pairs := strings.Split(fd, "&")
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

	if bf.jar != nil {
		bf.client.Jar = bf.jar
	}

	r := request.(BaseFetcherRequest)
	var err error
	var req *http.Request
	var resp *http.Response

	if request.GetFormData() == "" {
		req, err = http.NewRequest(r.Method, r.URL, nil)
		if err != nil {
			return nil, err
		}
	} else {
		//if form data exists send POST request
		formData := parseFormData(r.GetFormData())
		req, err = http.NewRequest("POST", r.URL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
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
		case 401:
			return nil, &errs.Unauthorized{}
		case 407:
			return nil, &errs.ProxyAuthenticationRequired{}
		case 500:
			return nil, &errs.InternalServerError{}
		case 504:
			return nil, &errs.GatewayTimeout{}
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
	return &response, err
}

func (bf *BaseFetcher) GetCookieJar() *cookiejar.Jar {
	return bf.jar
}

func (bf *BaseFetcher) SetCookieJar(jar *cookiejar.Jar) {
	bf.jar = jar
}

// Static type assertion
var _ Fetcher = &BaseFetcher{}
