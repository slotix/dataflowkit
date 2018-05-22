package fetch

import (
	"io"
	"net/url"

	"github.com/spf13/viper"

	"github.com/PuerkitoBio/goquery"

	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/utils"
)

// ProxyJSCSSMiddleware replace css/js href/src to use proxy
func ProxyJSCSSMiddleware() ServiceMiddleware {
	return func(next Service) Service {
		return JSCSSReplaceMiddleware{next}
	}
}

type JSCSSReplaceMiddleware struct {
	Service
}

//Response increments requst count before sending it to actual Response service handler.
func (mw JSCSSReplaceMiddleware) Response(req FetchRequester) (response FetchResponser, err error) {
	response, err = mw.Service.Response(req)
	if err != nil {
		return nil, err
	}
	content, err := response.GetHTML()
	if err != nil {
		return nil, err
	}
	respURL := response.GetURL()
	html, err := mw.replaceHref(content, respURL)
	if err != nil {
		return nil, err
	}

	var fetchResponse FetchResponser
	switch req.Type() {
	case "base":
		fetchResponse = &BaseFetcherResponse{
			Expires:           response.GetExpires(),
			URL:               response.GetURL(),
			ReasonsNotToCache: response.GetReasonsNotToCache(),
			HTML:              html} //response.(*BaseFetcherResponse)
	case "splash":
		fetchResponse = &splash.Response{
			Expires:           response.GetExpires(),
			URL:               response.GetURL(),
			ReasonsNotToCache: response.GetReasonsNotToCache(),
			HTML:              html} //response.(*splash.Response)
	}

	return fetchResponse, nil
}

func (mw JSCSSReplaceMiddleware) replaceHref(content io.Reader, baseURL string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(content)
	if err != nil {
		return "", err
	}
	doc.Find("link,a").Each(func(_ int, sel *goquery.Selection) {
		href, ok := sel.Attr("href")
		if ok {
			href, err = utils.RelUrl(baseURL, href)
			if err != nil {
				logger.Error(err)
			}

			if goquery.NodeName(sel) == "link" {
				href = viper.GetString("DFK_PROXY_HOST") + "/proxy?url=" + url.PathEscape(href)
			}
			sel.SetAttr("href", href)
		}
	})
	doc.Find("script,img").Each(func(_ int, sel *goquery.Selection) {
		src, ok := sel.Attr("src")
		if ok {
			src, err = utils.RelUrl(baseURL, src)
			if err != nil {
				logger.Error(err)
			}
			if goquery.NodeName(sel) == "script" {
				src = viper.GetString("DFK_PROXY_HOST") + "/proxy?url=" + url.PathEscape(src)
			}
			sel.SetAttr("src", src)
		}
	})
	doc.Find("form").Each(func(_ int, sel *goquery.Selection) {
		action, ok := sel.Attr("action")
		if ok {
			action, err = utils.RelUrl(baseURL, action)
			if err != nil {
				logger.Error(err)
			}
			sel.SetAttr("action", action)
		}
	})
	str, err := goquery.OuterHtml(doc.Selection)
	if err != nil {
		return "", err
	}
	return str, nil
}
