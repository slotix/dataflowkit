package scrape

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/slotix/dataflowkit/errs"
)

type dummyPaginator struct {
}

func (p dummyPaginator) NextPage(uri string, doc *goquery.Selection) (string, error) {
	return "", nil
}

func attrOrDataValue(s *goquery.Selection) (value string) {
	if s.Length() == 0 {
		return ""
	}
	attr, exists := s.Attr("class")
	attr = strings.TrimSpace(attr)
	if exists && attr != "" { //in some cases tag is invalid f.e. <tr class>
		var re = regexp.MustCompile(`\n?\s{1,}`)
		attr = "." + re.ReplaceAllString(attr, `.`)
		return attr
	}
	attr, exists = s.Attr("id")

	if exists && attr != "" {
		return fmt.Sprintf("#%s", attr)
	}
	//if len(s.Nodes)>0 {
	return s.Nodes[0].Data
	//}
}

// DividePageByIntersection returns DividePageFunc function
// which determines common ancestor of specified selectors.
func DividePageByIntersection(selectors []string) DividePageFunc {
	ret := func(doc *goquery.Selection) []*goquery.Selection {
		sels := []*goquery.Selection{}
		sel, err := getCommonAncestor(doc, selectors)
		if err != nil {
			// no common ancestor returned
			return nil
		}

		sel.Each(func(i int, s *goquery.Selection) {
			sels = append(sels, s)
		})

		return sels
	}
	return ret
}

func getCommonAncestor(doc *goquery.Selection, selectors []string) (*goquery.Selection, error) {
	selectorAncestor := doc.Find(selectors[0]).First().Parent()
	if len(selectors) > 1 {
		bFound := false
		selectorsSlice := selectors[1:]
		for !bFound {
			for _, f := range selectorsSlice {
				sel := doc.Find(f).First()
				sel = sel.ParentsUntilSelection(selectorAncestor).Last()
				//check last node.. if it equal html its mean that first selector's parent
				//not found
				if goquery.NodeName(sel) == "html" {
					selectorAncestor = doc.FindSelection(selectorAncestor.Parent().First())
					bFound = false
					break
				}
				bFound = true
			}
		}
	}
	if selectorAncestor.Length() == 0 {
		return nil, &errs.BadPayload{errs.ErrNoCommonAncestor}
	}
	fullPath := goquery.NodeName(selectorAncestor)
	parents := selectorAncestor.ParentsUntilSelection(doc.Find("body"))
	parents.Each(func(i int, s *goquery.Selection) {
		//avoid antiscrapin' tech like twitter
		selector := attrOrDataValue(s)
		fullPath = selector + " > " + fullPath
	})
	items := doc.Find(fullPath)
	return items, nil
}
