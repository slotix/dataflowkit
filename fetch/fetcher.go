package fetch

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mafredri/cdp/rpcc"
	"github.com/spf13/viper"
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

func isExclude(origin string) bool {
	excludeRes := viper.GetStringSlice("EXCLUDERES")
	for _, res := range excludeRes {
		if strings.Index(origin, res) != -1 {
			return true
		}
	}
	return false
}

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
