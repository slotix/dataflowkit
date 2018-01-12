package extract

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"regexp"
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

func TestText(t *testing.T) {
	sel := selFrom(`<p>Test 123</p>`)
	ret, err := Text{}.Extract(sel)
	assert.NoError(t, err)
	assert.Equal(t, ret, "Test 123")

	sel = selFrom(`<p>First</p><p>Second</p>`)
	ret, err = Text{}.Extract(sel)
	assert.NoError(t, err)
	assert.Equal(t, ret, "FirstSecond")

	sel = selFrom(`<p>First</p><p>Second</p><p>Third</p>`)
	ret, err = Text{}.Extract(sel.Find("p"))
	assert.NoError(t, err)
	assert.Equal(t, ret, []string{"First", "Second", "Third"})
}

func TestHtml(t *testing.T) {
	sel := selFrom(
		`<div class="one">` +
			`<div class="two">Bar</div>` +
			`<div class="two"><i>Baz</i></div>` +
			`<div class="three">Asdf</div>` +
			`</div>`)
	ret, err := Html{}.Extract(sel.Find(".one"))
	assert.NoError(t, err)
	assert.Equal(t, ret, `<div class="two">Bar</div><div class="two"><i>Baz</i></div><div class="three">Asdf</div>`)

	ret, err = Html{}.Extract(sel.Find(".two"))
	assert.NoError(t, err)
	assert.Equal(t, ret, `Bar<i>Baz</i>`)
}

func TestOuterHtml(t *testing.T) {
	// Simple version
	sel := selFrom(`<div><p>Test 123</p></div>`)
	ret, err := OuterHtml{}.Extract(sel.Find("p"))
	assert.NoError(t, err)
	assert.Equal(t, ret, `<p>Test 123</p>`)

	// Should only get the outer HTML of the element, not siblings
	sel = selFrom(`<div><p>Test 123</p><b>foo</b></div>`)
	ret, err = OuterHtml{}.Extract(sel.Find("p"))
	assert.NoError(t, err)
	assert.Equal(t, ret, `<p>Test 123</p>`)
}

func TestRegexInvalid(t *testing.T) {
	var err error

	_, err = Regex{}.Extract(selFrom(`foo`))
	assert.Error(t, err, "no regex given")

	_, err = Regex{Regex: regexp.MustCompile(`foo`)}.Extract(selFrom(`bar`))
	assert.Error(t, err, "regex has no subexpressions")

	_, err = Regex{Regex: regexp.MustCompile(`(a)(b)`)}.Extract(selFrom(`bar`))
	assert.Error(t, err, "regex has more than one subexpression (2), but which to extract was not specified")
}

func TestRegex(t *testing.T) {
	sel := selFrom(`<div class="one">foo</div><div class="fooobar">bar</div>`)
	ret, err := Regex{Regex: regexp.MustCompile("f(o+)o")}.Extract(sel)
	assert.NoError(t, err)
	assert.Equal(t, ret, []string{"o", "oo"})

	ret, err = Regex{
		Regex:         regexp.MustCompile("f(o)?(oo)bar"),
		Subexpression: 2,
	}.Extract(sel)
	assert.NoError(t, err)
	assert.Equal(t, ret, "oo")

	ret, err = Regex{
		Regex:    regexp.MustCompile("f(o+)o"),
		OnlyText: true,
	}.Extract(sel)
	assert.NoError(t, err)
	assert.Equal(t, ret, "o")

	ret, err = Regex{
		Regex:            regexp.MustCompile("f(o+)o"),
		OnlyText:         true,
		AlwaysReturnList: true,
	}.Extract(sel)
	assert.NoError(t, err)
	assert.Equal(t, ret, []string{"o"})

	ret, err = Regex{
		Regex:          regexp.MustCompile("a(sd)f"),
		IncludeIfEmpty: false,
	}.Extract(sel)
	assert.NoError(t, err)
	assert.Nil(t, ret)
}

func TestAttrInvalid(t *testing.T) {
	var err error

	_, err = Attr{}.Extract(selFrom(`foo`))
	assert.Error(t, err, "no attribute provided")
}

func TestAttr(t *testing.T) {
	sel := selFrom(`
	<a href="http://www.google.com">google</a>
	<a href="http://www.yahoo.com">yahoo</a>
	<a href="http://www.microsoft.com" class="notsearch">microsoft</a>
	`)

	ret, err := Attr{Attr: "href"}.Extract(sel.Find("a"))
	assert.NoError(t, err)
	assert.Equal(t, ret, []string{
		"http://www.google.com",
		"http://www.yahoo.com",
		"http://www.microsoft.com",
	})

	ret, err = Attr{Attr: "href"}.Extract(sel.Find(".notsearch"))
	assert.NoError(t, err)
	assert.Equal(t, ret, "http://www.microsoft.com")

	ret, err = Attr{Attr: "href", AlwaysReturnList: true}.Extract(sel.Find(".notsearch"))
	assert.NoError(t, err)
	assert.Equal(t, ret, []string{"http://www.microsoft.com"})

	// ret, err = Attr{
	// 	Attr:             "href",
	// 	AlwaysReturnList: true,
	// }.Extract(sel.Find(".abc"))
	// assert.NoError(t, err)
	// assert.Equal(t, ret, []string{})

	ret, err = Attr{
		Attr:           "href",
		IncludeIfEmpty: false,
	}.Extract(sel.Find(".abc"))
	assert.NoError(t, err)
	assert.Nil(t, ret)
}

func TestImgAttr(t *testing.T) {
	sel := selFrom(`
	<img src="smiley.gif" alt="Smiley face" height="42" width="42">
	`)
	ret, err := Attr{Attr: "src"}.Extract(sel.Find("img"))
	assert.NoError(t, err)
	assert.Equal(t, ret, string("smiley.gif"))
	ret, err = Attr{Attr: "alt"}.Extract(sel.Find("img"))
	assert.NoError(t, err)
	assert.Equal(t, ret, string("Smiley face"))

}

func TestLink(t *testing.T) {
	sel := selFrom(`
		<a href="http://www.google.com">google</a>
		<a href="http://www.yahoo.com">yahoo</a>
		<a href="http://www.microsoft.com" class="notsearch">microsoft</a>
		`)

	ret, err := Link{}.Extract(sel.Find("a"))
	assert.NoError(t, err)
	m := []map[string]string{{"google": "http://www.google.com"}, {"yahoo": "http://www.yahoo.com"}, {"microsoft": "http://www.microsoft.com"}}

	assert.Equal(t, ret, m, "check if maps are equal")
}
func TestCount(t *testing.T) {
	sel := selFrom(`
	<div>One</div>
	<div class="foo">Two</div>
	<div>Three</div>
	`)

	ret, err := Count{}.Extract(sel.Find("div"))
	assert.NoError(t, err)
	assert.Equal(t, ret, 3)

	ret, err = Count{}.Extract(sel.Find(".foo"))
	assert.NoError(t, err)
	assert.Equal(t, ret, 1)

	// ret, err = Count{}.Extract(sel.Find(".bad"))
	// assert.NoError(t, err)
	// assert.Equal(t, ret, 0)

	ret, err = Count{IncludeIfEmpty: false}.Extract(sel.Find(".bad"))
	assert.NoError(t, err)
	assert.Nil(t, ret)
}
