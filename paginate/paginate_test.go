package paginate

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func selFrom(s string) *goquery.Selection {
	r := strings.NewReader(s)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		panic(err)
	}

	return doc.Selection
}

func TestBySelector(t *testing.T) {
	sel := selFrom(`<a href="http://www.google.com">foo</a>`)

	pg, err := BySelector("a", "href").NextPage("", sel)
	assert.NoError(t, err)
	assert.Equal(t, pg, "http://www.google.com")

	pg, err = BySelector("div", "xxx").NextPage("", sel)
	assert.NoError(t, err)
	assert.Equal(t, pg, "")

	sel = selFrom(`<a href="/foobar">foo</a>`)

	pg, err = BySelector("a", "href").NextPage("http://www.google.com", sel)
	assert.NoError(t, err)
	assert.Equal(t, pg, "http://www.google.com/foobar")

	sel = selFrom(`<a href="asdf?q=123">foo</a>`)

	pg, err = BySelector("a", "href").NextPage("http://www.google.com", sel)
	assert.NoError(t, err)
	assert.Equal(t, pg, "http://www.google.com/asdf?q=123")
}

func TestByQueryParam(t *testing.T) {
	pg, err := ByQueryParam("foo").NextPage("http://www.google.com?foo=1", nil)
	assert.NoError(t, err)
	assert.Equal(t, pg, "http://www.google.com?foo=2")

	pg, err = ByQueryParam("bad").NextPage("http://www.google.com", nil)
	assert.NoError(t, err)
	assert.Equal(t, pg, "")

	pg, err = ByQueryParam("bad").NextPage("http://www.google.com?bad=asdf", nil)
	assert.NoError(t, err)
	assert.Equal(t, pg, "")
}
