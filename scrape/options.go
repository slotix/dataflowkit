package scrape

// ScrapeOptions contains options that are used during the progress of a
// scrape.
type ScrapeOptions struct {
	// The maximum number of pages to scrape.  The scrape will proceed until
	// either this number of pages have been scraped, or until the paginator
	// returns no further URLs.  Set this value to 0 to indicate an unlimited
	// number of pages can be scraped.
	MaxPages int
	//Output format 
	Format string
	//paginated results are returned. Single list of combined results from every block on all pages is returned by default. Paginated results is actual for JSON and XML formats. Combined list of results is always returned for CSV format.  
	PaginatedResults bool
}	

// The default options during a scrape.
var DefaultOptions = ScrapeOptions{
	MaxPages: 0,
	Format: "json",
	PaginatedResults: false,
}
