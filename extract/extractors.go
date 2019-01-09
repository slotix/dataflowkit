package extract

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"bytes"
	"errors"
	"regexp"

	"github.com/slotix/dataflowkit/utils"

	"github.com/PuerkitoBio/goquery"

	"golang.org/x/net/html"
)

// The Extractor interface represents something that can extract data from
// a selection.
type Extractor interface {
	// Extract some data from the given Selection and return it.  The returned
	// data should be encodable - i.e. passing it to json.Marshal should succeed.
	// If the returned data is nil, then the output from this part will not be
	// included.
	//
	// If this function returns an error, then the scrape is aborted.
	Extract(*goquery.Selection) (interface{}, error)
}

// Const is an Extractor that returns a constant value.
type Const struct {
	// The value to return when the Extract() function is called.
	Val interface{}
}

// Extract returns Const value.
func (e Const) Extract(sel *goquery.Selection) (interface{}, error) {
	return e.Val, nil
}

var _ Extractor = Const{}

// Text is an Extractor that returns the combined text contents of
// the given selection.
type Text struct {
	// If text is empty in the selection, then return the empty string from Extract,
	// instead of 'nil'.  This signals that the result of this Part
	// should be included to the results, as opposed to omitting the
	// empty string.
	IncludeIfEmpty bool
	//Filters are used to manipulate Text data when extracting.
	//Currently the following filters are available:
	//upperCase makes all of the letters in the selected text  uppercase.
	//lowerCase makes all of the letters in the selected text lowercase.
	//capitalize capitalizes the first letter of each word in the selected text
	//trim returns a copy of the text, with all leading and trailing white space removed
	Filters []string
}

// Extract returns Text value from specified selection.
func (e Text) Extract(sel *goquery.Selection) (interface{}, error) {
	results := []string{}
	sel.Each(func(i int, s *goquery.Selection) {
		//filtering text data.
		filtered := filterText(s.Text(), e.Filters)
		results = append(results, filtered)
	})

	if len(results) == 0 && !e.IncludeIfEmpty {
		return nil, nil
	}

	if len(results) == 1 {
		return results[0], nil
	}

	return results, nil
}

var _ Extractor = Text{}

// Html extracts and returns the HTML from inside each part of the
// given selection, as a string.
//
// Note that this results in what is effectively the innerHTML of the element -
// i.e. if our selection consists of ["<p><b>ONE</b></p>", "<p><i>TWO</i></p>"]
// then the output will be: "<b>ONE</b><i>TWO</i>".
//
// The return type is a string of all the inner HTML joined together.
//type Html struct{}

// Extract returns HTML from specified selection.
// func (e Html) Extract(sel *goquery.Selection) (interface{}, error) {
// 	var ret, h string
// 	var err error

// 	sel.EachWithBreak(func(i int, s *goquery.Selection) bool {
// 		h, err = s.Html()
// 		if err != nil {
// 			return false
// 		}

// 		ret += h
// 		return true
// 	})

// 	if err != nil {
// 		return nil, err
// 	}
// 	return ret, nil
// }

// var _ Extractor = Html{}

// OuterHtml extracts and returns the HTML of each part of the
// given selection, as a string.
//
// To illustrate, if our selection consists of
// ["<div><b>ONE</b></div>", "<p><i>TWO</i></p>"] then the output will be:
// "<div><b>ONE</b></div><p><i>TWO</i></p>".
//
// The return type is a string of all the outer HTML joined together.
type OuterHtml struct{}

// Extract returns OuterHtml from specified selection.
func (e OuterHtml) Extract(sel *goquery.Selection) (interface{}, error) {
	output := bytes.NewBufferString("")
	for _, node := range sel.Nodes {
		if err := html.Render(output, node); err != nil {
			return nil, err
		}
	}

	return output.String(), nil
}

var _ Extractor = OuterHtml{}

// Regex runs the given regex over the contents of each part in the
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
	// each part in the selection, as opposed to the HTML contents.
	OnlyText bool

	// By default, if there is only a single match, Regex will return
	// the match itself (as opposed to an array containing the single match).
	// Set AlwaysReturnList to true to disable this behaviour, ensuring that the
	// Extract function always returns an array.
	AlwaysReturnList bool

	// If no matches of the provided regex could be extracted, then return the empty list
	// from Extract, instead of 'nil'.  This signals that the result of
	// this Part should be included to the results, as opposed to
	// omitting the empty list.
	IncludeIfEmpty bool
}

// Extract returns Regex'ed  value from specified selection.
func (e Regex) Extract(sel *goquery.Selection) (interface{}, error) {
	//if e.Regex == nil {
	//	return nil, errors.New("no regex given")
	//}
	//if e.Regex.NumSubexp() == 0 {
	//	return nil, errors.New("regex has no subexpressions")
	//}
	//
	//var subexp int
	//if e.Subexpression == 0 {
	//	if e.Regex.NumSubexp() != 1 {
	//		e := fmt.Errorf(
	//			"regex has more than one subexpression (%d), but which to "+
	//				"extract was not specified",
	//			e.Regex.NumSubexp())
	//		return nil, e
	//	}
	//
	//	subexp = 1
	//} else {
	//	subexpЕ = e.Subexpression
	//}
	var subexp int
	if subexp = e.Subexpression; subexp == 0 {
		subexp = 1
	}
	results := []string{}

	// For each part in the selector...
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

var _ Extractor = Regex{}

// Attr extracts the value of a given HTML attribute from each part
// in the selection, and returns them as a list.
// The return type of the extractor is a list of attribute values (i.e. []string).
type Attr struct {
	// The HTML attribute to extract from each part.
	Attr string
	//BaseURL specifies the base URL to use for all relative URLs contained within a document.
	BaseURL string
	// By default, if there is only a single attribute extracted, AttrExtractor
	// will return the match itself (as opposed to an array containing the single
	// match). Set AlwaysReturnList to true to disable this behaviour, ensuring
	// that the Extract function always returns an array.
	AlwaysReturnList bool

	// If no parts with this attribute are found, then return the empty list from
	// Extract, instead of  'nil'.  This signals that the result of this
	// Part should be included to the results, as opposed to omitting
	// the empty list.
	IncludeIfEmpty bool
	//Filters are used to manipulate HTML attribute when extracting.
	//Currently the following filters are available:
	//upperCase makes all of the letters in the Attr uppercase.
	//lowerCase makes all of the letters in the Attr lowercase.
	//capitalizes the first letter of each word in the Attr
	//trim returns a copy of the Attr, with all leading and trailing white space removed
	Filters []string
}

// Extract returns Attr value from specified selection.
//Absolute URL will be returned for href and src attributes if relative URLs provided
func (e Attr) Extract(sel *goquery.Selection) (interface{}, error) {
	if len(e.Attr) == 0 {
		return nil, errors.New("no attribute provided")
	}
	results := []string{}
	sel.Each(func(i int, s *goquery.Selection) {
		if val, found := s.Attr(e.Attr); found {
			if e.Attr == "href" || e.Attr == "src" {
				var err error
				//transform relative url to absolute url
				val, err = utils.RelUrl(e.BaseURL, val)
				if err != nil {
					logger.Error(err.Error())
				}
			}
			filtered := filterText(val, e.Filters)
			results = append(results, filtered)
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

var _ Extractor = Attr{}

// Count extracts the count of parts that are matched and returns it.
type Count struct {
	// If no parts with this attribute are found, then return a number from
	// Extract, instead of 'nil'.  This signals that the result of this
	// Part should be included to the results, as opposed to omitting
	// the empty list.
	IncludeIfEmpty bool
}

// Extract returns length of elements in selection.
func (e Count) Extract(sel *goquery.Selection) (interface{}, error) {
	l := sel.Length()
	if l == 0 && !e.IncludeIfEmpty {
		return nil, nil
	}

	return l, nil
}

var _ Extractor = Count{}
