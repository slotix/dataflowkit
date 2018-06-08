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
	if exists && attr != "" { //in some cases tag is invalid f.e. <tr class>
		attr = strings.TrimSpace(attr)
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

// func intersectionFL(sel *goquery.Selection) *goquery.Selection {
// 	first := sel.First()
// 	last := sel.Last()
// 	intersection := first.Parents().Intersection(last.Parents())
// 	return intersection
// }

// func findIntersection(doc *goquery.Selection, selectors []string) (*goquery.Selection, error) {
// 	var intersection *goquery.Selection
// 	for i, f := range selectors {
// 		sel := doc.Find(f)
// 		//logger.Info(f, sel.Length())
// 		//col.genAttrFieldName(f.Name, sel)
// 		if sel.Length() > 0 { //don't add selectors to intersection if length is 0. Otherwise the whole intersection returns No selectors error
// 			if i == 0 {
// 				intersection = intersectionFL(sel)
// 			} else {
// 				intersection = intersection.Intersection(intersectionFL(sel))
// 			}
// 		}
// 	}
// 	//logger.Info(attrOrDataValue(intersection))
// 	if intersection == nil || intersection.Length() == 0 {
// 		return nil, &errs.BadPayload{errs.ErrNoSelectors}
// 	}
// 	intersectionWithParent := fmt.Sprintf("%s>%s",
// 		attrOrDataValue(intersection.Parent()),
// 		attrOrDataValue(intersection))
// 	//logger.Info(intersectionWithParent)
// 	items := doc.Find(intersectionWithParent)
// 	//return intersectionWithParent, nil
// 	//logger.Info(items.Length())

// 	var inter1 *goquery.Selection
// 	if items.Length() == 1 {
// 		inter1 = items.Children()
// 		//sel = fmt.Sprintf("%s>%s>%s",
// 		//	attrOrDataValue(intersection.Parent()),
// 		//	attrOrDataValue(intersection),
// 		//	attrOrDataValue(intersection.Children()))

// 	}
// 	if items.Length() > 1 {
// 		inter1 = items
// 		//sel = intersectionWithParent
// 	}
// 	return inter1, nil
// }

// DividePageBySelector returns a function that divides a page into blocks by
// CSS selector.  Each element in the page with the given selector is treated
// as a new block.
// func DividePageBySelector(sel string) DividePageFunc {
// 	ret := func(doc *goquery.Selection) []*goquery.Selection {
// 		sels := []*goquery.Selection{}
// 		doc.Find(sel).Each(func(i int, s *goquery.Selection) {
// 			sels = append(sels, s)
// 		})

// 		return sels
// 	}
// 	return ret
// }

// DividePageByIntersection returns DividePageFunc function
// which determines common ancestor of specified selectors.
func DividePageByIntersection(selectors []string) DividePageFunc {
	ret := func(doc *goquery.Selection) []*goquery.Selection {
		sels := []*goquery.Selection{}
		//doc.Find(sel).Each(func(i int, s *goquery.Selection) {
		sel, err := getCommonAncestor(doc, selectors)
		//sel, err = findIntersection(doc, selectors)
		if err != nil {
			logger.Warn(err)
			return nil
		}

		sel.Each(func(i int, s *goquery.Selection) {
			sels = append(sels, s)
			//	logger.Println(attrOrDataValue(s))
		})

		return sels
	}
	return ret
}

func getCommonAncestor(doc *goquery.Selection, selectors []string) (*goquery.Selection, error) {
	// if len(selectors) == 0 {
	// 	return nil, &errs.BadPayload{errs.ErrNoSelectors}
	// }
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
