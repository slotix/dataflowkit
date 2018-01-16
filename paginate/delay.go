package paginate

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"time"

	"github.com/PuerkitoBio/goquery"
)

type withDelayPaginator struct {
	delay time.Duration
	p     Paginator
}

// WithDelay returns a Paginator that will wait the given duration whenever the
// next page is requested, and will then dispatch to the underling Paginator.
func WithDelay(delay time.Duration, p Paginator) Paginator {
	return &withDelayPaginator{
		delay: delay,
		p:     p,
	}
}

func (p *withDelayPaginator) NextPage(uri string, doc *goquery.Selection) (string, error) {
	time.Sleep(p.delay)
	return p.p.NextPage(uri, doc)
}
