package scrape

import (
	"github.com/PuerkitoBio/goquery"
)

type dummyPaginator struct {
}

func (p dummyPaginator) NextPage(uri string, doc *goquery.Selection) (string, error) {
	return "", nil
}

// DividePageBySelector returns a function that divides a page into blocks by
// CSS selector.  Each element in the page with the given selector is treated
// as a new block.
func DividePageBySelector(sel string) DividePageFunc {
	ret := func(doc *goquery.Selection) []*goquery.Selection {
		sels := []*goquery.Selection{}
		doc.Find(sel).Each(func(i int, s *goquery.Selection) {
			sels = append(sels, s)
		})

		return sels
	}
	return ret
}
