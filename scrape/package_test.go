package scrape_test

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/andrew-d/goscrape"
	"github.com/slotix/dataflowkit/extract"
	"github.com/stretchr/testify/assert"
)

func TestDefaultPaginator(t *testing.T) {
	sc := mustNew(&scrape.ScrapeConfig{
		Fetcher: newDummyFetcher([][]byte{
			[]byte("one"),
			[]byte("two"),
			[]byte("three"),
			[]byte("four"),
		}),

		Pieces: []scrape.Piece{
			{Name: "dummy", Selector: ".", Extractor: extract.Const{"asdf"}},
		},
	})

	results, err := sc.ScrapeWithOpts(
		"initial",
		scrape.ScrapeOptions{MaxPages: 3},
	)
	assert.NoError(t, err)
	assert.Equal(t, results.URLs, []string{"initial"})
	assert.Equal(t, len(results.Results), 1)
	assert.Equal(t, len(results.Results[0]), 1)
}

func TestPageLimits(t *testing.T) {
	sc := mustNew(&scrape.ScrapeConfig{
		Fetcher: newDummyFetcher([][]byte{
			[]byte("one"),
			[]byte("two"),
			[]byte("three"),
			[]byte("four"),
		}),

		Paginator: &dummyPaginator{},

		Pieces: []scrape.Piece{
			{Name: "dummy", Selector: ".", Extractor: extract.Const{"asdf"}},
		},
	})

	results, err := sc.ScrapeWithOpts(
		"initial",
		scrape.ScrapeOptions{MaxPages: 3},
	)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"initial",
		"url-1",
		"url-2",
	}, results.URLs)
}

func mustNew(c *scrape.ScrapeConfig) *scrape.Scraper {
	scraper, err := scrape.New(c)
	if err != nil {
		panic(err)
	}
	return scraper
}

type dummyFetcher struct {
	data [][]byte
	idx  int
}

func newDummyFetcher(data [][]byte) *dummyFetcher {
	return &dummyFetcher{
		data: data,
		idx:  0,
	}
}

func (d *dummyFetcher) Prepare() error {
	return nil
}

func (d *dummyFetcher) Fetch(method, url string) (io.ReadCloser, error) {
	r := dummyReadCloser{bytes.NewReader(d.data[d.idx])}
	d.idx++
	return r, nil
}

func (d *dummyFetcher) Close() {
	return
}

type dummyPaginator struct {
	idx int
}

func (d *dummyPaginator) NextPage(url string, document *goquery.Selection) (string, error) {
	d.idx++
	return fmt.Sprintf("url-%d", d.idx), nil
}

type dummyReadCloser struct {
	u io.Reader
}

func (d dummyReadCloser) Read(b []byte) (int, error) {
	return d.u.Read(b)
}

func (d dummyReadCloser) Close() error {
	return nil
}
