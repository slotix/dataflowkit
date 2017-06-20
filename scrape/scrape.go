package scrape

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/paginate"
	"github.com/slotix/dataflowkit/splash"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "scrape: ", log.Lshortfile)
}

var (
	ErrNoPieces = errors.New("no pieces in the config")
)

// The DividePageFunc type is used to extract a page's blocks during a scrape.
// For more information, please see the documentation on the ScrapeConfig type.
type DividePageFunc func(*goquery.Selection) []*goquery.Selection

// A Piece represents a given chunk of data that is to be extracted from every
// block in each page of a scrape.
type Piece struct {
	// The name of this piece.  Required, and will be used to aggregate results.
	Name string

	// A sub-selector within the given block to process.  Pass in "." to use
	// the root block's selector with no modification.
	Selector string
	// TODO(andrew-d): Consider making this an interface too.

	// Extractor contains the logic on how to extract some results from the
	// selector that is provided to this Piece.
	Extractor extract.PieceExtractor
}

// The main configuration for a scrape.  Pass this to the New() function.
type ScrapeConfig struct {
	// Fetcher is the underlying transport that is used to fetch documents.
	// If this is not specified (i.e. left nil), then a default SplashFetcher
	// will be created and used.
	Fetcher Fetcher

	// Paginator is the Paginator to use for this current scrape.
	//
	// If Paginator is nil, then no pagination is performed and it is assumed that
	// the initial URL is the only page.
	Paginator paginate.Paginator

	// DividePage splits a page into individual 'blocks'.  When scraping, we treat
	// each page as if it contains some number of 'blocks', each of which can be
	// further subdivided into what actually needs to be extracted.
	//
	// If the DividePage function is nil, then no division is performed and the
	// page is assumed to contain a single block containing the entire <body>
	// element.
	DividePage DividePageFunc

	// Pieces contains the list of data that is extracted for each block.  For
	// every block that is the result of the DividePage function (above), all of
	// the Pieces entries receives the selector representing the block, and can
	// return a result.  If the returned result is nil, then the Piece is
	// considered not to exist in this block, and is not included.
	//
	// Note: if a Piece's Extractor returns an error, it results in the scrape
	// being aborted - this can be useful if you need to ensure that a given Piece
	// is required, for example.
	Pieces []Piece

	Opts ScrapeOptions
}

/*
func (c *ScrapeConfig) clone() *ScrapeConfig {
	ret := &ScrapeConfig{
		Fetcher:    c.Fetcher,
		Paginator:  c.Paginator,
		DividePage: c.DividePage,
		Pieces:     c.Pieces,
	}
	return ret
}
*/

// ScrapeResults describes the results of a scrape.  It contains a list of all
// pages (URLs) visited during the process, along with all results generated
// from each Piece in each page.
type ScrapeResults struct {
	// All URLs visited during this scrape, in order.  Always contains at least
	// one element - the initial URL.
	URLs []string

	// The results from each Piece of each page.  Essentially, the top-level array
	// is for each page, the second-level array is for each block in a page, and
	// the final map[string]interface{} is the mapping of Piece.Name to results.
	Results [][]map[string]interface{}
}

// First returns the first set of results - i.e. the results from the first
// block on the first page.
// This function can return nil if there were no blocks found on the first page
// of the scrape.
func (r *ScrapeResults) First() map[string]interface{} {
	if len(r.Results[0]) == 0 {
		return nil
	}

	return r.Results[0][0]
}

// AllBlocks returns a single list of results from every block on all pages.
// This function will always return a list, even if no blocks were found.
func (r *ScrapeResults) AllBlocks() []map[string]interface{} {
	ret := []map[string]interface{}{}

	for _, page := range r.Results {
		for _, block := range page {
			ret = append(ret, block)
		}
	}

	return ret
}

type Scraper struct {
	config *ScrapeConfig
}

// Create a new scraper with the provided configuration.
func New(c *ScrapeConfig) (*Scraper, error) {
	var err error

	// Validate config
	if len(c.Pieces) == 0 {
		return nil, ErrNoPieces
	}

	seenNames := map[string]struct{}{}
	for i, piece := range c.Pieces {
		if len(piece.Name) == 0 {
			return nil, fmt.Errorf("no name provided for piece %d", i)
		}
		if _, seen := seenNames[piece.Name]; seen {
			return nil, fmt.Errorf("piece %s has a duplicate name", i)
		}
		seenNames[piece.Name] = struct{}{}

		if len(piece.Selector) == 0 {
			return nil, fmt.Errorf("no selector provided for piece %d", i)
		}
	}

	// Clone the configuration and fill in the defaults.
	config := c //.clone()
	if config.Paginator == nil {
		config.Paginator = dummyPaginator{}
	}
	if config.DividePage == nil {
		config.DividePage = DividePageBySelector("body")

	}

	if config.Fetcher == nil {
		//config.Fetcher, err = NewHttpClientFetcher()
		config.Fetcher, err = NewSplashFetcher()
		if err != nil {
			return nil, err
		}
	}

	// All set!
	ret := &Scraper{
		config: config,
	}
	return ret, nil
}

// Scrape a given URL with default options.  See 'ScrapeWithOpts' for more
// information.
func (s *Scraper) Scrape(req interface{}) (*ScrapeResults, error) {
	return s.ScrapeWithOpts(req, DefaultOptions)
}

// Actually start scraping at the given URL.
//
// Note that, while this function and the Scraper in general are safe for use
// from multiple goroutines, making multiple requests in parallel can cause
// strange behaviour - e.g. overwriting cookies in the underlying http.Client.
// Please be careful when running multiple scrapes at a time, unless you know
// that it's safe.
func (s *Scraper) ScrapeWithOpts(req interface{}, opts ScrapeOptions) (*ScrapeResults, error) {

	var url string
	switch v := req.(type) {
	case HttpClientFetcherRequest:
		url = v.URL
	case splash.Request:
		url = v.URL
	}
	logger.Println(req)
	//rt := fmt.Sprintf("%T\n", req)
	//logger.Println(r)

	if len(url) == 0 {
		return nil, errors.New("no URL provided")
	}

	// Prepare the fetcher.
	err := s.config.Fetcher.Prepare()
	if err != nil {
		return nil, err
	}

	res := &ScrapeResults{
		URLs:    []string{},
		Results: [][]map[string]interface{}{},
	}

	var numPages int
	for {
		// Repeat until we don't have any more URLs, or until we hit our page limit.
		if len(url) == 0 || (opts.MaxPages > 0 && numPages >= opts.MaxPages) {
			break
		}

		r, err := s.config.Fetcher.Fetch(req)

		if err != nil {
			return nil, err
		}

		var resp io.ReadCloser
		switch req.(type) {
		case HttpClientFetcherRequest:
			resp = r.(io.ReadCloser)
		case splash.Request:
			if sResponse, ok := r.(*splash.Response); ok {
				resp, err = sResponse.GetContent()
				if err != nil {
					logger.Println(err)
				}
			}
		}

		// Create a goquery document.
		doc, err := goquery.NewDocumentFromReader(resp)
		resp.Close()
		if err != nil {
			return nil, err
		}
		res.URLs = append(res.URLs, url)
		results := []map[string]interface{}{}

		// Divide this page into blocks
		for _, block := range s.config.DividePage(doc.Selection) {
			blockResults := map[string]interface{}{}

			// Process each piece of this block
			for _, piece := range s.config.Pieces {
				//logger.Println(piece)
				sel := block
				if piece.Selector != "." {
					sel = sel.Find(piece.Selector)
				}

				pieceResults, err := piece.Extractor.Extract(sel)
				//logger.Println(attrOrDataValue(sel))
				if err != nil {
					return nil, err
				}

				// A nil response from an extractor means that we don't even include it in
				// the results.
				//if pieceResults == nil || pieceResults == "" {
				if pieceResults == nil {
					continue
				}

				blockResults[piece.Name] = pieceResults
			}
			if len(blockResults) > 0 {
				// Append the results from this block.
				results = append(results, blockResults)
			}
		}

		// Append the results from this page.
		res.Results = append(res.Results, results)

		numPages++

		// Get the next page.
		url, err = s.config.Paginator.NextPage(url, doc.Selection)
		if err != nil {
			return nil, err
		}

		switch req.(type) {
		case HttpClientFetcherRequest:
			req = HttpClientFetcherRequest{URL: url}
		case splash.Request:
			//every time when getting a response the next request will be filled with updated cookie information
			if sResponse, ok := r.(*splash.Response); ok {
				setCookie, err := sResponse.SetCookieToRequest()
				if err != nil {
					//return nil, err
					logger.Println(err)
				}
				req = splash.Request{URL: url, Cookies: setCookie}
			}
		}
	}

	// All good!
	return res, nil
}
