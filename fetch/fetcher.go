package fetch

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
	"github.com/slotix/dataflowkit/errs"
	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/sync/errgroup"
)

//Type represents types of fetcher
type Type string

//Fetcher types
const (
	//Base fetcher is used for downloading html web page using Go standard library's http
	Base Type = "Base"
	//Headless chrome is used to download content from JS driven web pages
	Chrome = "Chrome"
)

// Fetcher is the interface that must be satisfied by things that can fetch
// remote URLs and return their contents.
//
// Note: Fetchers may or may not be safe to use concurrently.  Please read the
// documentation for each fetcher for more details.
type Fetcher interface {
	//  Fetch is called to retrieve HTML content of a document from the remote server.
	Fetch(request Request) (io.ReadCloser, error)
	getCookieJar() http.CookieJar
	setCookieJar(jar http.CookieJar)
	getCookies(u *url.URL) ([]*http.Cookie, error)
	setCookies(u *url.URL, cookies []*http.Cookie) error
}

//Request struct contains request information sent to  Fetchers
type Request struct {
	// Type defines Fetcher type. It may be "chrome" or "base". Defaults to "base". 
	Type string `json:"type"`
	//	URL to be retrieved
	URL string `json:"url"`
	//	HTTP method : GET, POST
	Method string
	// FormData is a string value for passing formdata parameters.
	//
	// For example it may be used for processing pages which require authentication
	//
	// Example:
	//
	// "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=userpassword&rememberMe=1"
	//
	FormData string `json:"formData,omitempty"`
	//UserToken identifies user to keep personal cookies information.
	UserToken string `json:"userToken"`
	// Actions contains the list of action we have to perform on page
	Actions string `json:"actions"`
}

// BaseFetcher is a Fetcher that uses the Go standard library's http
// client to fetch URLs.
type BaseFetcher struct {
	client *http.Client
}

// ChromeFetcher is used to fetch Java Script rendeded pages.
type ChromeFetcher struct {
	cdpClient *cdp.Client
	client    *http.Client
	cookies   []*http.Cookie
}

//newFetcher creates instances of Fetcher for downloading a web page.
func newFetcher(t Type) Fetcher {
	switch t {
	case Base:
		return newBaseFetcher()
	case Chrome:
		return newChromeFetcher()
	default:
		logger.Panic(fmt.Sprintf("unhandled type: %#v", t))
	}
	panic("unreachable")
}

// newBaseFetcher creates instances of newBaseFetcher{} to fetch
// a page content from regular websites as-is
// without running js scripts on the page.
func newBaseFetcher() *BaseFetcher {
	var client *http.Client
	proxy := viper.GetString("PROXY")
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			logger.Error(err.Error())
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
	jarOpts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	var err error
	f.client.Jar, err = cookiejar.New(jarOpts)
	if err != nil {
		return nil
	}
	return f
}

// Fetch retrieves document from the remote server. It returns web page content along with cache and expiration information.
func (bf *BaseFetcher) Fetch(request Request) (io.ReadCloser, error) {
	resp, err := bf.response(request)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

//Response return response after document fetching using BaseFetcher
func (bf *BaseFetcher) response(r Request) (*http.Response, error) {
	//URL validation
	if _, err := url.ParseRequestURI(r.getURL()); err != nil {
		return nil, err
	}
	var err error
	var req *http.Request

	if r.FormData == "" {
		req, err = http.NewRequest(r.Method, r.URL, nil)
		if err != nil {
			return nil, err
		}
	} else {
		//if form data exists send POST request
		formData := parseFormData(r.FormData)
		req, err = http.NewRequest("POST", r.URL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	}
	//TODO: Add UA to requests
	//req.Header.Add("User-Agent", "Dataflow kit - https://github.com/slotix/dataflowkit")
	return bf.doRequest(req)
}

func (bf *BaseFetcher) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := bf.client.Do(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
		return resp, nil

	default:
		return nil, errs.StatusError{
			resp.StatusCode,
			errors.New(http.StatusText(resp.StatusCode)),
		}
	}
}

func (bf *BaseFetcher) getCookieJar() http.CookieJar { //*cookiejar.Jar {
	return bf.client.Jar
}

func (bf *BaseFetcher) setCookieJar(jar http.CookieJar) {

	bf.client.Jar = jar
}

func (bf *BaseFetcher) getCookies(u *url.URL) ([]*http.Cookie, error) {
	return bf.client.Jar.Cookies(u), nil
}

func (bf *BaseFetcher) setCookies(u *url.URL, cookies []*http.Cookie) error {
	bf.client.Jar.SetCookies(u, cookies)
	return nil
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

// Static type assertion
var _ Fetcher = &BaseFetcher{}

// NewChromeFetcher returns ChromeFetcher
func newChromeFetcher() *ChromeFetcher {
	var client *http.Client
	proxy := viper.GetString("PROXY")
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			logger.Error(err.Error())
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

// LogCodec captures the output from writing RPC requests and reading
// responses on the connection. It implements rpcc.Codec via
// WriteRequest and ReadResponse.
type LogCodec struct{ conn io.ReadWriter }

// WriteRequest marshals v into a buffer, writes its contents onto the
// connection and logs it.
func (c *LogCodec) WriteRequest(req *rpcc.Request) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return err
	}
	fmt.Printf("SEND: %s", buf.Bytes())
	_, err := c.conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// ReadResponse unmarshals from the connection into v whilst echoing
// what is read into a buffer for logging.
func (c *LogCodec) ReadResponse(resp *rpcc.Response) error {
	var buf bytes.Buffer
	if err := json.NewDecoder(io.TeeReader(c.conn, &buf)).Decode(resp); err != nil {
		return err
	}
	fmt.Printf("RECV: %s\n", buf.String())
	return nil
}

// Fetch retrieves document from the remote server. It returns web page content along with cache and expiration information.
func (f *ChromeFetcher) Fetch(request Request) (io.ReadCloser, error) {
	//URL validation
	if _, err := url.ParseRequestURI(strings.TrimSpace(request.getURL())); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	devt := devtool.New(viper.GetString("CHROME"), devtool.WithClient(f.client))
	//https://github.com/mafredri/cdp/issues/60
	//pt, err := devt.Get(ctx, devtool.Page)
	pt, err := devt.Create(ctx)
	if err != nil {
		return nil, err
	}
	var conn *rpcc.Conn
	if viper.GetBool("CHROME_TRACE") {
		newLogCodec := func(conn io.ReadWriter) rpcc.Codec {
			return &LogCodec{conn: conn}
		}
		// Connect to WebSocket URL (page) that speaks the Chrome Debugging Protocol.
		conn, err = rpcc.DialContext(ctx, pt.WebSocketDebuggerURL, rpcc.WithCodec(newLogCodec))
	} else {
		conn, err = rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	}
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer conn.Close() // Cleanup.
	defer devt.Close(ctx, pt)
	// Create a new CDP Client that uses conn.
	f.cdpClient = cdp.NewClient(conn)

	if err = runBatch(
		// Enable all the domain events that we're interested in.
		func() error { return f.cdpClient.DOM.Enable(ctx) },
		func() error { return f.cdpClient.Network.Enable(ctx, nil) },
		func() error { return f.cdpClient.Page.Enable(ctx) },
		func() error { return f.cdpClient.Runtime.Enable(ctx) },
	); err != nil {
		return nil, err
	}

	err = f.loadCookies()
	if err != nil {
		return nil, err
	}
	domLoadTimeout := 60 * time.Second
	if request.FormData == "" {
		err = f.navigate(ctx, f.cdpClient.Page, "GET", request.getURL(), "", domLoadTimeout)
	} else {
		formData := parseFormData(request.FormData)
		err = f.navigate(ctx, f.cdpClient.Page, "POST", request.getURL(), formData.Encode(), domLoadTimeout)
	}
	if err != nil {
		return nil, err
	}

	if err := f.runActions(ctx, request.Actions); err != nil {
		logger.Warn(err.Error())
	}

	u, err := url.Parse(request.getURL())
	if err != nil {
		return nil, err
	}
	f.cookies, err = f.saveCookies(u)
	if err != nil {
		return nil, err
	}

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

func (f *ChromeFetcher) runActions(ctx context.Context, actionsJSON string) error {
	acts := []map[string]json.RawMessage{}
	err := json.Unmarshal([]byte(actionsJSON), &acts)
	if err != nil {
		return err
	}
	for _, actionMap := range acts {
		for actionType, params := range actionMap {
			action, err := NewAction(actionType, params)
			if err == nil {
				return action.Execute(ctx, f)
			}
		}
	}
	return nil
}

func (f *ChromeFetcher) setCookieJar(jar http.CookieJar) {
	f.client.Jar = jar
}

func (f *ChromeFetcher) getCookieJar() http.CookieJar {
	return f.client.Jar
}

// Static type assertion
var _ Fetcher = &ChromeFetcher{}

// navigate to the URL and wait for DOMContentEventFired. An error is
// returned if timeout happens before DOMContentEventFired.
func (f *ChromeFetcher) navigate(ctx context.Context, pageClient cdp.Page, method, url string, formData string, timeout time.Duration) error {
	defer time.Sleep(750 * time.Millisecond)

	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), timeout)

	// Make sure Page events are enabled.
	err := pageClient.Enable(ctxTimeout)
	if err != nil {
		return err
	}

	// Navigate to GitHub, block until ready.
	loadEventFired, err := pageClient.LoadEventFired(ctxTimeout)
	if err != nil {
		return err
	}
	defer loadEventFired.Close()

	loadingFailed, err := f.cdpClient.Network.LoadingFailed(ctxTimeout)
	if err != nil {
		return err
	}
	defer loadingFailed.Close()

	// exceptionThrown, err := f.cdpClient.Runtime.ExceptionThrown(ctxTimeout)
	// if err != nil {
	// 	return err
	// }
	//defer exceptionThrown.Close()

	if method == "GET" {
		_, err = pageClient.Navigate(ctxTimeout, page.NewNavigateArgs(url))
		if err != nil {
			return err
		}
	} else {
		/* ast := "*" */
		pattern := network.RequestPattern{URLPattern: &url}
		patterns := []network.RequestPattern{pattern}

		f.cdpClient.Network.SetCacheDisabled(ctxTimeout, network.NewSetCacheDisabledArgs(true))

		interArgs := network.NewSetRequestInterceptionArgs(patterns)
		err = f.cdpClient.Network.SetRequestInterception(ctxTimeout, interArgs)
		if err != nil {
			return err
		}

		kill := make(chan bool)
		go f.interceptRequest(ctxTimeout, url, formData, kill)
		_, err = pageClient.Navigate(ctxTimeout, page.NewNavigateArgs(url))
		if err != nil {
			return err
		}
		kill <- true
	}
	select {
	// case <-exceptionThrown.Ready():
	// 	ev, err := exceptionThrown.Recv()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return errs.StatusError{400, errors.New(ev.ExceptionDetails.Error())}
	case <-loadEventFired.Ready():
		_, err = loadEventFired.Recv()
		if err != nil {
			return err
		}
	case <-loadingFailed.Ready():
		reply, err := loadingFailed.Recv()
		if err != nil {
			return err
		}
		if reply.Type == network.ResourceTypeDocument {
			return errs.StatusError{400, errors.New(reply.ErrorText)}
		}
	case <-ctx.Done():
		cancelTimeout()
		return nil /*
			case <-ctxTimeout.Done():
				return errs.StatusError{400, errors.New("Fetch timeout")} */
	}
	return nil
}

func (f *ChromeFetcher) setCookies(u *url.URL, cookies []*http.Cookie) error {
	f.cookies = cookies
	return nil
}

func (f *ChromeFetcher) loadCookies() error {
	/* 	u, err := url.Parse(cookiesURL)
	   	if err != nil {
	   		return err
	   	} */
	for _, c := range f.cookies {
		c1 := network.SetCookieArgs{
			Name:  c.Name,
			Value: c.Value,
			Path:  &c.Path,
			/* Expires:  expire, */
			Domain:   &c.Domain,
			HTTPOnly: &c.HttpOnly,
			Secure:   &c.Secure,
		}
		if !c.Expires.IsZero() {
			duration := c.Expires.Sub(time.Unix(0, 0))
			c1.Expires = network.TimeSinceEpoch(duration / time.Second)
		}
		_, err := f.cdpClient.Network.SetCookie(context.Background(), &c1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *ChromeFetcher) getCookies(u *url.URL) ([]*http.Cookie, error) {
	return f.cookies, nil
}

func (f *ChromeFetcher) saveCookies(u *url.URL) ([]*http.Cookie, error) {
	ncookies, err := f.cdpClient.Network.GetCookies(context.Background(), &network.GetCookiesArgs{URLs: []string{u.String()}})
	if err != nil {
		return nil, err
	}
	cookies := []*http.Cookie{}
	for _, c := range ncookies.Cookies {

		c1 := http.Cookie{
			Name:  c.Name,
			Value: c.Value,
			Path:  c.Path,
			/* Expires:  expire, */
			Domain:   c.Domain,
			HttpOnly: c.HTTPOnly,
			Secure:   c.Secure,
		}
		if c.Expires > -1 {
			sec, dec := math.Modf(c.Expires)
			expire := time.Unix(int64(sec), int64(dec*(1e9)))
			/* logger.Info(expire.String())
			logger.Info(expire.Format("2006-01-02 15:04:05")) */
			c1.Expires = expire
		}
		cookies = append(cookies, &c1)
	}
	return cookies, nil
}

func (f *ChromeFetcher) interceptRequest(ctx context.Context, originURL string, formData string, kill chan bool) {
	var sig = false
	cl, err := f.cdpClient.Network.RequestIntercepted(ctx)
	if err != nil {
		panic(err)
	}
	defer cl.Close()
	for {
		if sig {
			return
		}
		select {
		case <-cl.Ready():
			r, err := cl.Recv()
			if err != nil {
				logger.Error(err.Error())
				sig = true
				continue
			}

			if len(formData) > 0 && r.Request.URL == originURL && r.RedirectURL == nil {
				interceptedArgs := network.NewContinueInterceptedRequestArgs(r.InterceptionID)
				interceptedArgs.SetMethod("POST")
				interceptedArgs.SetPostData(formData)
				fData := fmt.Sprintf(`{"Content-Type":"application/x-www-form-urlencoded","Content-Length":%d}`, len(formData))
				interceptedArgs.Headers = []byte(fData)
				if err = f.cdpClient.Network.ContinueInterceptedRequest(ctx, interceptedArgs); err != nil {
					logger.Error(err.Error())
					sig = true
					continue
				}
			} else {
				interceptedArgs := network.NewContinueInterceptedRequestArgs(r.InterceptionID)
				if r.ResourceType == network.ResourceTypeImage || r.ResourceType == network.ResourceTypeStylesheet || isExclude(r.Request.URL) {
					interceptedArgs.SetErrorReason(network.ErrorReasonAborted)
				}
				if err = f.cdpClient.Network.ContinueInterceptedRequest(ctx, interceptedArgs); err != nil {
					logger.Error(err.Error())
					sig = true
					continue
				}
				continue
			}
		case <-kill:
			sig = true
			break
		}
	}
}

func isExclude(origin string) bool {
	excludeRes := viper.GetStringSlice("EXCLUDERES")
	for _, res := range excludeRes {
		if strings.Index(origin, res) != -1 {
			return true
		}
	}
	return false
}

func (f ChromeFetcher) RunJSFromFile(ctx context.Context, path string, entryPointFunction string) error {
	exp, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	exp = append(exp, entryPointFunction...)

	compileReply, err := f.cdpClient.Runtime.CompileScript(ctx, &runtime.CompileScriptArgs{
		Expression:    string(exp),
		PersistScript: true,
	})
	if err != nil {
		panic(err)
	}
	awaitPromise := true

	_, err = f.cdpClient.Runtime.RunScript(ctx, &runtime.RunScriptArgs{
		ScriptID:     *compileReply.ScriptID,
		AwaitPromise: &awaitPromise,
	})
	return err
}

// removeNodes deletes all provided nodeIDs from the DOM.
// func removeNodes(ctx context.Context, domClient cdp.DOM, nodes ...dom.NodeID) error {
// 	var rmNodes []runBatchFunc
// 	for _, id := range nodes {
// 		arg := dom.NewRemoveNodeArgs(id)
// 		rmNodes = append(rmNodes, func() error { return domClient.RemoveNode(ctx, arg) })
// 	}
// 	return runBatch(rmNodes...)
// }

// runBatchFunc is the function signature for runBatch.
type runBatchFunc func() error

// runBatch runs all functions simultaneously and waits until
// execution has completed or an error is encountered.
func runBatch(fn ...runBatchFunc) error {
	eg := errgroup.Group{}
	for _, f := range fn {
		eg.Go(f)
	}
	return eg.Wait()
}

//GetURL returns URL to be fetched
func (req Request) getURL() string {
	return strings.TrimRight(strings.TrimSpace(req.URL), "/")
}

// Host returns Host value from Request
func (req Request) Host() (string, error) {
	u, err := url.Parse(req.getURL())
	if err != nil {
		return "", err
	}
	return u.Host, nil
}
