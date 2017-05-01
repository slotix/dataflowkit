package scrape

// ScrapeOptions contains options that are used during the progress of a
// scrape.
type ScrapeOptions struct {
	// The maximum number of pages to scrape.  The scrape will proceed until
	// either this number of pages have been scraped, or until the paginator
	// returns no further URLs.  Set this value to 0 to indicate an unlimited
	// number of pages can be scraped.
	MaxPages int
}

// The default options during a scrape.
var DefaultOptions = ScrapeOptions{
	MaxPages: 0,
}
