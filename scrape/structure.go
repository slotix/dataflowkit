package scrape

import (
	"io"
	"sync"
	"time"

	"github.com/slotix/dataflowkit/storage"

	"github.com/slotix/dataflowkit/fetch"

	"github.com/PuerkitoBio/goquery"
	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/paginate"
	"github.com/temoto/robotstxt"
)

// Extractor type represents Extractor types available for scraping.
// Here is the list of Extractor types are currently supported:
// text, html, outerHtml, attr, link, image, regex, const, count
// Find more actual information in docs/extractors.md
type Extractor struct {
	Types []string `json:"types"`
	// Params are unique for each type
	Params  map[string]interface{} `json:"params"`
	Filters []string               `json:"filters"`
}

//A Field corresponds to a given chunk of data to be extracted from every block in each page of a scrape.
type Field struct {
	//Name is a name of fields. It is required, and will be used to aggregate results.
	Name string `json:"name"`
	//Selector is a CSS selector within the given block to process.  Pass in "." to use the root block's selector.
	Selector string `json:"selector"`
	//Extractor contains the logic on how to extract some results from the selector that is provided to this Field.
	Extractor Extractor `json:"extractor"`
	//Details is an optional field strictly for Link extractor type. It guides scraper to parse additional pages following the links according to the set of fields specified inside "details"
	Details *details `json:"details"`
}

type details struct {
	Fields    []Field    `json:"fields"`
	Paginator *paginator `json:"paginator"`
	IsPath    bool       `json:"path"`
}

// paginator is used to scrape multiple pages.
// paginator extracts the next page from a document by querying a given CSS selector and extracting the given HTML attribute from the resulting element.
type paginator struct {
	//Selector represents CSS selector for the next page
	Selector string `json:"selector"`
	// HTML attribute for the next page
	Attribute string `json:"attr"`
	// The maximum number of pages to scrape. The scrape will proceed until either this number of pages have been scraped, or until the paginator returns no further URLs.
	//
	// Default value is 1.
	// Set this value to 0 to indicate an unlimited number of pages to be scraped.
	//
	MaxPages       int  `json:"maxPages"`
	InfiniteScroll bool `json:"infiniteScroll"`
}

// Payload structure contain information and rules to be passed to a scraper
// Find the most actual information in docs/payload.md
type Payload struct {
	// Name - Collection name.
	Name string `json:"name"`
	//Request struct represents HTTP request to be sent to a server. It combines parameters for passing for downloading html pages by Fetch Endpoint.
	//Request.URL field is required. All other fields including Params, Cookies, Func are optional.
	Request fetch.FetchRequester `json:"request"`
	//Fields is a set of fields used to extract data from a web page.
	Fields []Field `json:"fields"`
	//PayloadMD5 encodes payload content to MD5. It is used for generating file name to be stored.
	PayloadMD5 []byte `json:"payloadMD5"`
	//FetcherType represent fetcher which is used for document download.
	//Set up it to either `base` or `splash` values
	//If FetcherType is omited the value of FETCHER_TYPE of parse.d service is used by default.
	FetcherType string `json:"fetcherType"`
	//Format represents output format (CSV, JSON, XML)
	Format string `json:"format"`
	//Paginator is used to scrape multiple pages.
	//If Paginator is nil, then no pagination is performed and it is assumed that the initial URL is the only page.
	Paginator *paginator `json:"paginator"`
	//Paginated results are returned if true.
	//Default value is false
	// Single list of combined results from every block on all pages is returned by default.
	//
	// Paginated results are applicable for JSON and XML output formats.
	//
	// Combined list of results is always returned for CSV format.
	PaginateResults *bool `json:"paginateResults"`
	//FetchDelay should be used for a scraper to throttle the crawling speed to avoid hitting the web servers too frequently.
	//FetchDelay specifies sleep time for multiple requests for the same domain. It is equal to FetchDelay * random value between 500 and 1500 msec
	FetchDelay *time.Duration `json:"fetchDelay"`
	//Some web sites track  statistically significant similarities in the time between requests to them. RandomizeCrawlDelay setting decreases the chance of a crawler being blocked by such sites. This way a random delay ranging from 0.5 * CrawlDelay to 1.5 * CrawlDelay seconds is used between consecutive requests to the same domain. If CrawlDelay is zero (default) this option has no effect.
	RandomizeFetchDelay *bool `json:"randomizeFetchDelay"`
	//Maximum number of times to retry, in addition to the first download.
	//RETRY_HTTP_CODES
	//Default: [500, 502, 503, 504, 408]
	//Failed pages should be rescheduled for download at the end. once the spider has finished crawling all other (non failed) pages.
	RetryTimes int `json:"retryTimes"`
	// ContainPath means that one of the field just a path and we have to ignore all other fields (if present)
	// that are not a path
	IsPath bool `json:"path"`
}

// The DividePageFunc type is used to extract a page's blocks during a scrape.
// For more information, please see the documentation on the ScrapeConfig type.
type DividePageFunc func(*goquery.Selection) []*goquery.Selection

// A Part represents a given chunk of data that is to be extracted from every
// block in each page of a scrape.
type Part struct {
	// The name of this part.  Required, and will be used to aggregate results.
	Name string

	// A sub-selector within the given block to process.  Pass in "." to use
	// the root block's selector with no modification.
	Selector string

	// Extractor contains the logic on how to extract some results from the
	// selector that is provided to this Piece.
	Extractor extract.Extractor
	//Details is an optional field strictly for Link extractor type. It guides scraper to parse additional pages following the links according to the set of fields specified inside "details"
	Details Scraper
}

//Scraper struct consolidates settings for scraping task.
type Scraper struct {
	Request fetch.FetchRequester
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
	// tag.
	DividePage DividePageFunc

	// Parts contains the list of data that is extracted for each block.  For
	// every block that is the result of the DividePage function (above), all of
	// the Parts entries receives the selector representing the block, and can
	// return a result.  If the returned result is nil, then the Part is
	// considered not to exist in this block, and is not included.
	//
	// Note: if a Part's Extractor returns an error, it results in the scrape
	// being aborted - this can be useful if you need to ensure that a given Part
	// is required, for example.
	Parts []Part
	//Opts contains options that are used during the progress of a
	// scrape.
	//Opts ScrapeOptions
	IsPath bool
}

// Results describes the results of a scrape.  It contains a list of all
// pages (URLs) visited during the process, along with all results generated
// from each Part in each page.
type Results struct {

	// Output represents combined results after parsing from each Part of each page.  Essentially, the top-level array
	// is for each page, the second-level array is for each block in a page, and
	// the final map[string]interface{} is the mapping of Part.Name to results.
	Output [][]map[string]interface{}
}

// Task keeps Results of Task generated from Payload along with other auxiliary information
type Task struct {
	ID      string
	Payload Payload
	//Scrapers []*Scraper
	// Visited contain a map[url]error during this scrape.
	// Always contains at least one element - the initial URL.
	//Failed pages should be rescheduled for download at the end if during a scrape one of the following statuses returned [500, 502, 503, 504, 408]
	//once the spider has finished crawling all other (non failed) pages.
	Errors []error
	//TaskQueue chan *Scraper
	Robots map[string]*robotstxt.RobotsData
	//Results
	Parsed bool
	// Block counter
	BlockCounter uint32
}

type worker struct {
	wg      *sync.WaitGroup
	scraper *Scraper
	storage *storage.Store
	mx      *sync.Mutex
}

type taskWorker struct {
	wg              *sync.WaitGroup
	UID             string
	currentPageNum  int
	keys            *map[int]uint32
	scraper         *Scraper
	storage         *storage.Store
	mx              *sync.Mutex
	useBlockCounter bool
}

type blockStruct struct {
	blockSelection  *goquery.Selection
	key             string
	hash            string
	useBlockCounter bool
}

type fetchInfo struct {
	result  chan<- io.ReadCloser
	request fetch.FetchRequester
	err     chan<- error
}
