package fetch

import (
	"fmt"
	"context"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
	"golang.org/x/sync/errgroup"
)

//ChromeFetcherRequest struct contains request information used by ChromeFetcher
type ChromeFetcherRequest struct {
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
}

//	GetFormData returns form data from ChromeFetcherRequest
func (req ChromeFetcherRequest) GetFormData() string {
	return req.FormData
}

//GetURL returns URL to be fetched
func (req ChromeFetcherRequest) GetURL() string {
	return strings.TrimRight(strings.TrimSpace(req.URL), "/")
}

//GetUserToken returns user token
func (r ChromeFetcherRequest) GetUserToken() string {
	return r.UserToken
}

// Host returns Host value from Request
func (req ChromeFetcherRequest) Host() (string, error) {
	u, err := url.Parse(req.GetURL())
	if err != nil {
		return "", err
	}
	return u.Host, nil
}

//Type returns fetcher type
func (req ChromeFetcherRequest) Type() string {
	return "chrome"
}

// setCookies sets all the provided cookies.
// func setCookies(ctx context.Context, net cdp.Network, cookies ...Cookie) error {
// 	var cmds []runBatchFunc
// 	for _, c := range cookies {
// 		args := network.NewSetCookieArgs(c.Name, c.Value).SetURL(c.URL)
// 		cmds = append(cmds, func() error {
// 			reply, err := net.SetCookie(ctx, args)
// 			if err != nil {
// 				return err
// 			}
// 			if !reply.Success {
// 				return errors.New("could not set cookie")
// 			}
// 			return nil
// 		})
// 	}
// 	return runBatch(cmds...)
// }

// navigate to the URL and wait for DOMContentEventFired. An error is
// returned if timeout happens before DOMContentEventFired.
func (f *ChromeFetcher) navigate(ctx context.Context, pageClient cdp.Page, method, url string, formData string, timeout time.Duration) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// Make sure Page events are enabled.
	err := pageClient.Enable(ctx)
	if err != nil {
		return err
	}

	// Open client for DOMContentEventFired to block until DOM has fully loaded.
	domContentEventFired, err := pageClient.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContentEventFired.Close()

	if method == "GET" {
		_, err = pageClient.Navigate(ctx, page.NewNavigateArgs(url))
		if err != nil {
			return err
		}
	} else {
		go func() {
			cl, err := f.cdpClient.Network.RequestIntercepted(ctx)
			r, err := cl.Recv()
			if err != nil {
				panic(err)
			}
			interceptedArgs := network.NewContinueInterceptedRequestArgs(r.InterceptionID)
			interceptedArgs.SetMethod("POST")
			interceptedArgs.SetPostData(formData)
			fData := fmt.Sprintf(`{"Content-Type":"application/x-www-form-urlencoded","Content-Length":%d}`,len(formData))
			interceptedArgs.Headers = []byte(fData)
			if err = f.cdpClient.Network.ContinueInterceptedRequest(ctx, interceptedArgs); err != nil {
				panic(err)
			}
		}()
		_, err = pageClient.Navigate(ctx, page.NewNavigateArgs(url))
		if err != nil {
			return err
		}
	}
	_, err = domContentEventFired.Recv()
	return err
}

func (f ChromeFetcher) runJSFromFile(ctx context.Context, path string) error {
	exp, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	compileReply, err := f.cdpClient.Runtime.CompileScript(context.Background(), &runtime.CompileScriptArgs{
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
func removeNodes(ctx context.Context, domClient cdp.DOM, nodes ...dom.NodeID) error {
	var rmNodes []runBatchFunc
	for _, id := range nodes {
		arg := dom.NewRemoveNodeArgs(id)
		rmNodes = append(rmNodes, func() error { return domClient.RemoveNode(ctx, arg) })
	}
	return runBatch(rmNodes...)
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
