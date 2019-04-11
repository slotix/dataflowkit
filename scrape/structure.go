package scrape

import (
	"sync"
	"time"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/storage"
	"github.com/temoto/robotstxt"
)

//A Field corresponds to a given chunk of data to be extracted from every block in each page of a scrape.
type Field struct {
	//Name is a name of fields. It is required, and will be used to aggregate results.
	Name string `json:"name"`
	//Selector is a CSS selector within the given block to process.  Pass in "." to use the root block's selector.
	CSSSelector string `json:"selector"`
	//Attrs specify attributes which will be extracted from element
	Attrs []string `json:"attrs"`
	//Details is an optional field strictly for Link extractor type. It guides scraper to parse additional pages following the links according to the set of fields specified inside "details"
	Details Payload `json:"details"`
	//Filters
	Filters []Filter `json:"filters"`
}

type Filter struct {
	Name  string
	Param string
}

// Payload structure contain information and rules to be passed to a scraper
// Find the most actual information in docs/payload.md
type Payload struct {
	// Name - Collection name.
	Name string `json:"name"`
	//Request struct represents HTTP request to be sent to a server. It combines parameters for passing for downloading html pages by Fetch Endpoint.
	//Request.URL field is required. All other fields including Params, Cookies, Func are optional.
	Request fetch.Request `json:"request"`
	//Fields is a set of fields used to extract data from a web page.
	Fields []Field `json:"fields"`
	//PayloadMD5 encodes payload content to MD5. It is used for generating file name to be stored.
	PayloadMD5 string
	//FetcherType represent fetcher which is used for document download.
	//Set up it to either `base` or `chrome` values
	//If FetcherType is omitted the value of FETCHER_TYPE of parse.d service is used by default.
	//FetcherType string `json:"fetcherType"`
	//Format represents output format (CSV, JSON, XML)
	Format string `json:"format"`
	//Compressed represents if result will be compressed into GZip
	Compressor string `json:"compressor"`
	//Paginator is used to scrape multiple pages.
	//If Paginator is nil, then no pagination is performed and it is assumed that the initial URL is the only page.
	Paginator string `json:"paginator"`
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
	FetchDelay *time.Duration
	//Some web sites track  statistically significant similarities in the time between requests to them. RandomizeCrawlDelay setting decreases the chance of a crawler being blocked by such sites. This way a random delay ranging from 0.5  CrawlDelay to 1.5  CrawlDelay seconds is used between consecutive requests to the same domain. If CrawlDelay is zero (default) this option has no effect.
	RandomizeFetchDelay *bool
	//Maximum number of times to retry, in addition to the first download.
	//RETRY_HTTP_CODES
	//Default: [500, 502, 503, 504, 408]
	//Failed pages should be rescheduled for download at the end. once the spider has finished crawling all other (non failed) pages.
	RetryTimes int `json:"retryTimes"`
	// ContainPath means that one of the field just a path and we have to ignore all other fields (if present)
	// that are not a path
	IsPath       bool `json:"path"`
	blockCounter *int
}

// Task keeps Results of Task generated from Payload along with other auxiliary information
type Task struct {
	Robots map[string]*robotstxt.RobotsData
	// storage using to write result into corresponding storage type
	storage       storage.Store
	requestCount  int
	responseCount int

	jobDone  sync.WaitGroup
	payloads chan Payload

	// path stuffs
	rootUID         string
	mx              sync.Mutex
	templateRequest fetch.Request

	isParsed bool
}

type encodeInfo struct {
	payloadMD5    string
	extension     string
	compressor    string
	compressLevel int
	fieldNames    []string
}

type flow struct {
	key  string
	url  string
	data interface{}
}
