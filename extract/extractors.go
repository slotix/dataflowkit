package extract

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "extractor: ", log.Lshortfile)
}

// The PieceExtractor interface represents something that can extract data from
// a selection.
type PieceExtractor interface {
	// Extract some data from the given Selection and return it.  The returned
	// data should be encodable - i.e. passing it to json.Marshal should succeed.
	// If the returned data is nil, then the output from this piece will not be
	// included.
	//
	// If this function returns an error, then the scrape is aborted.
	Extract(*goquery.Selection) (interface{}, error)
}

// Const is a PieceExtractor that returns a constant value.
type Const struct {
	// The value to return when the Extract() function is called.
	Val interface{}
}

func (e Const) Extract(sel *goquery.Selection) (interface{}, error) {
	return e.Val, nil
}

var _ PieceExtractor = Const{}

// Text is a PieceExtractor that returns the combined text contents of
// the given selection.
type Text struct {
	// If text is empty in the selection, then return the empty string from Extract,
	// instead of 'nil'.  This signals that the result of this Piece
	// should be included to the results, as opposed to omitting the
	// empty string.
	IncludeIfEmpty bool
}

func (e Text) Extract(sel *goquery.Selection) (interface{}, error) {
	results := []string{}

	sel.Each(func(i int, s *goquery.Selection) {
		results = append(results, s.Text())
	})

	if len(results) == 0 && !e.IncludeIfEmpty {
		return nil, nil
	}

	if len(results) == 1 {
		return results[0], nil
	}

	return results, nil
}

var _ PieceExtractor = Text{}

// Html extracts and returns the HTML from inside each element of the
// given selection, as a string.
//
// Note that this results in what is effectively the innerHTML of the element -
// i.e. if our selection consists of ["<p><b>ONE</b></p>", "<p><i>TWO</i></p>"]
// then the output will be: "<b>ONE</b><i>TWO</i>".
//
// The return type is a string of all the inner HTML joined together.
type Html struct{}

func (e Html) Extract(sel *goquery.Selection) (interface{}, error) {
	var ret, h string
	var err error

	sel.EachWithBreak(func(i int, s *goquery.Selection) bool {
		h, err = s.Html()
		if err != nil {
			return false
		}

		ret += h
		return true
	})

	if err != nil {
		return nil, err
	}
	return ret, nil
}

var _ PieceExtractor = Html{}

// OuterHtml extracts and returns the HTML of each element of the
// given selection, as a string.
//
// To illustrate, if our selection consists of
// ["<div><b>ONE</b></div>", "<p><i>TWO</i></p>"] then the output will be:
// "<div><b>ONE</b></div><p><i>TWO</i></p>".
//
// The return type is a string of all the outer HTML joined together.
type OuterHtml struct{}

func (e OuterHtml) Extract(sel *goquery.Selection) (interface{}, error) {
	output := bytes.NewBufferString("")
	for _, node := range sel.Nodes {
		if err := html.Render(output, node); err != nil {
			return nil, err
		}
	}

	return output.String(), nil
}

var _ PieceExtractor = OuterHtml{}

// Regex runs the given regex over the contents of each element in the
// given selection, and, for each match, extracts the given subexpression.
// The return type of the extractor is a list of string matches (i.e. []string).
type Regex struct {
	// The regular expression to match.  This regular expression must define
	// exactly one parenthesized subexpression (sometimes known as a "capturing
	// group"), which will be extracted.
	Regex *regexp.Regexp
	// The subexpression of the regex to match.  If this value is not set, and if
	// the given regex has more than one subexpression, an error will be thrown.
	Subexpression int

	// When OnlyText is true, only run the given regex over the text contents of
	// each element in the selection, as opposed to the HTML contents.
	OnlyText bool

	// By default, if there is only a single match, Regex will return
	// the match itself (as opposed to an array containing the single match).
	// Set AlwaysReturnList to true to disable this behaviour, ensuring that the
	// Extract function always returns an array.
	AlwaysReturnList bool

	// If no matches of the provided regex could be extracted, then return the empty list
	// from Extract, instead of 'nil'.  This signals that the result of
	// this Piece should be included to the results, as opposed to
	// omitting the empty list.
	IncludeIfEmpty bool
}

func (e Regex) Extract(sel *goquery.Selection) (interface{}, error) {
	if e.Regex == nil {
		return nil, errors.New("no regex given")
	}
	if e.Regex.NumSubexp() == 0 {
		return nil, errors.New("regex has no subexpressions")
	}

	var subexp int
	if e.Subexpression == 0 {
		if e.Regex.NumSubexp() != 1 {
			e := fmt.Errorf(
				"regex has more than one subexpression (%d), but which to "+
					"extract was not specified",
				e.Regex.NumSubexp())
			return nil, e
		}

		subexp = 1
	} else {
		subexp = e.Subexpression
	}

	results := []string{}

	// For each element in the selector...
	var err error
	sel.EachWithBreak(func(i int, s *goquery.Selection) bool {
		var contents string
		if e.OnlyText {
			contents = s.Text()
		} else {
			contents, err = s.Html()
			if err != nil {
				return false
			}
		}

		ret := e.Regex.FindAllStringSubmatch(contents, -1)

		// A return value of nil == no match
		if ret == nil {
			return true
		}

		// For each regex match...
		for _, submatches := range ret {
			// The 0th entry will be the match of the entire string.  The 1st
			// entry will be the first capturing group, which is what we want to
			// extract.
			if len(submatches) > 1 {
				results = append(results, submatches[subexp])
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}
	if len(results) == 0 && !e.IncludeIfEmpty {
		return nil, nil
	}
	if len(results) == 1 && !e.AlwaysReturnList {
		return results[0], nil
	}

	return results, nil
}

/*
func (e Regex) FillParams(m map[string]interface{}) error {
	err := FillStruct(m, &e)
	if err != nil {
		return err
	}
	regExp := m["regexp"]
	e.Regex = regexp.MustCompile(regExp.(string))
	logger.Println(e)
	return nil
}
*/
var _ PieceExtractor = Regex{}

// Attr extracts the value of a given HTML attribute from each element
// in the selection, and returns them as a list.
// The return type of the extractor is a list of attribute values (i.e. []string).
type Attr struct {
	// The HTML attribute to extract from each element.
	Attr string

	// By default, if there is only a single attribute extracted, AttrExtractor
	// will return the match itself (as opposed to an array containing the single
	// match). Set AlwaysReturnList to true to disable this behaviour, ensuring
	// that the Extract function always returns an array.
	AlwaysReturnList bool

	// If no elements with this attribute are found, then return the empty list from
	// Extract, instead of  'nil'.  This signals that the result of this
	// Piece should be included to the results, as opposed to omitting
	// the empty list.
	IncludeIfEmpty bool
}

func (e Attr) Extract(sel *goquery.Selection) (interface{}, error) {
	if len(e.Attr) == 0 {
		return nil, errors.New("no attribute provided")
	}

	results := []string{}

	sel.Each(func(i int, s *goquery.Selection) {
		if val, found := s.Attr(e.Attr); found {
			results = append(results, val)
		}
	})

	if len(results) == 0 && !e.IncludeIfEmpty {
		return nil, nil
	}
	if len(results) == 1 && !e.AlwaysReturnList {
		return results[0], nil
	}

	return results, nil
}

var _ PieceExtractor = Attr{}

// Count extracts the count of elements that are matched and returns it.
type Count struct {
	// If no elements with this attribute are found, then return a number from
	// Extract, instead of 'nil'.  This signals that the result of this
	// Piece should be included to the results, as opposed to omitting
	// the empty list.
	IncludeIfEmpty bool
}

func (e Count) Extract(sel *goquery.Selection) (interface{}, error) {
	l := sel.Length()
	if l == 0 && !e.IncludeIfEmpty {
		return nil, nil
	}

	return l, nil
}

var _ PieceExtractor = Count{}

/*
// Link is a PieceExtractor that returns the combined text contents and  Attr{Attr: "href"} of
// the given selection
type Link struct {
	// Piece should be included to the results, as opposed to omitting
	// the empty list.
	//IncludeIfEmpty bool
}

type LinkResult struct {
	text string
	href string
}

func (e Link) Extract(sel *goquery.Selection) (interface{}, error) {
	t, err := Text{}.Extract(sel)
	if err != nil {
		logger.Println(err)
	}
	a, err := Attr{Attr: "href"}.Extract(sel)
	if err != nil {
		logger.Println(err)
	}
	return LinkResult{t.(string), a.(string)}, nil
}

var _ PieceExtractor = Link{}
*/
/*
func FillParams(t string, m map[string]interface{}) (scrape.PieceExtractor, error) {
	//var err error

	logger.Println(t)
	var e scrape.PieceExtractor
	switch t {
	case "text":
		e = &Text{}
	case "attr":
		e = &Attr{}
	case "regex":
		//e = &Regex{}
		r := &Regex{}
		regExp := m["regexp"]
		r.Regex = regexp.MustCompile(regExp.(string))
		e = r
	}
	if m != nil {
		err := FillStruct(m, e)
		if err != nil {
			return nil, err
		}
	}
	return e, nil

}
*/
