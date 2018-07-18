package fetch

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
	"golang.org/x/sync/errgroup"
)

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
			fData := fmt.Sprintf(`{"Content-Type":"application/x-www-form-urlencoded","Content-Length":%d}`, len(formData))
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
