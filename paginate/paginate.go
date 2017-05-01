package paginate

import (
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/andrew-d/goscrape"
)

// RelUrl is a helper function that aids in calculating the absolute URL from a
// base URL and relative URL.
func RelUrl(base, rel string) (string, error) {
	baseUrl, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	relUrl, err := url.Parse(rel)
	if err != nil {
		return "", err
	}

	newUrl := baseUrl.ResolveReference(relUrl)
	return newUrl.String(), nil
}

type bySelectorPaginator struct {
	sel  string
	attr string
}

// BySelector returns a Paginator that extracts the next page from a document by
// querying a given CSS selector and extracting the given HTML attribute from the
// resulting element.
func BySelector(sel, attr string) scrape.Paginator {
	return &bySelectorPaginator{
		sel: sel, attr: attr,
	}
}

func (p *bySelectorPaginator) NextPage(uri string, doc *goquery.Selection) (string, error) {
	val, found := doc.Find(p.sel).Attr(p.attr)
	if !found {
		return "", nil
	}

	return RelUrl(uri, val)
}

type byQueryParamPaginator struct {
	param string
}

// ByQueryParam returns a Paginator that returns the next page from a document
// by incrementing a given query parameter.  Note that this will paginate
// infinitely - you probably want to specify a maximum number of pages to
// scrape by using the ScrapeWithOpts method.
func ByQueryParam(param string) scrape.Paginator {
	return &byQueryParamPaginator{param}
}

func (p *byQueryParamPaginator) NextPage(u string, _ *goquery.Selection) (string, error) {
	// Parse
	uri, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	// Parse query
	vals, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return "", err
	}

	// Find query param and increment.  If it doesn't exist, then we just stop.
	params, ok := vals[p.param]
	if !ok || len(params) < 1 {
		return "", nil
	}

	parsed, err := strconv.ParseUint(params[0], 10, 64)
	if err != nil {
		// TODO: should this be fatal?
		return "", nil
	}

	// Put everything back together
	params[0] = strconv.FormatUint(parsed+1, 10)
	vals[p.param] = params
	query := vals.Encode()
	uri.RawQuery = query
	return uri.String(), nil
}
