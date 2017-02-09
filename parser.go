package parser

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var errNoSelectors = errors.New("No selectors found")

//Parse parses payload json structure and generate Out to be serializes as JSON, XML, CSV, Excel
func (p *Payloads) Parse() (Out, error) {
	//parse input and fill Payload structure
	out := Out{}

	for _, collection := range p.Collections {
		content, err := GetHTML(collection.URL)
		if err != nil {
			return out, err
		}
		outItem, err := collection.parseItem(content)
		if err != nil {
			//return out, err
			fmt.Printf("\"%s:\" %s\n", outItem.Name, err)
		}
		out.Element = append(out.Element, outItem)
	}
	return out, nil
}


//genAttrFieldName generates field name according to attributes
func (o *outItem) genAttrFieldName(fieldName string, sel *goquery.Selection) {
	if _, exists := sel.Attr("href"); exists {
		o.Fields = append(o.Fields, fmt.Sprintf("%s_text",
			fieldName), fmt.Sprintf("%s_href", fieldName))
	} else if _, exists := sel.Attr("src"); exists {
		o.Fields = append(o.Fields, fmt.Sprintf("%s_src",
			fieldName), fmt.Sprintf("%s_alt", fieldName))
	} else {
		o.Fields = append(o.Fields, fieldName)
	}
}

type Item struct {
	value map[string]interface{}
}

//fillOutItem fills OutItem item values according to attributes
func (item *Item) fillOutItem(fieldName string, s *goquery.Selection) {

	if len(s.Nodes) > 0 && s.Nodes[0].Type == html.ElementNode {
		//fmt.Println("fillOut", s.Nodes[0].Data)
		nodeType := s.Nodes[0].Data
		//if href, exists := s.Attr("href"); exists {
		if nodeType == "a" {
			m := make(map[string]interface{})
			if href, exists := s.Attr("href"); exists {
				m["href"] = href
				m["text"] = strings.TrimSpace(s.Text())
				if title, exists := s.Attr("title"); exists {
					m["title"] = strings.TrimSpace(title)
				}
				item.value[fieldName] = m
			}
			//	} else if src, exists := s.Attr("src"); exists {
		} else if nodeType == "img" {

			m := make(map[string]interface{})
			if src, exists := s.Attr("src"); exists {
				m["src"] = src
				if alt, exists := s.Attr("alt"); exists {
					m["alt"] = strings.TrimSpace(alt)
				}
				//	m["width"] = strings.TrimSpace(s.AttrOr("width", ""))
				//	m["height"] = strings.TrimSpace(s.AttrOr("height", ""))
				item.value[fieldName] = m
			}
		} else {
			item.value[fieldName] = strings.TrimSpace(s.Text())
		}
	}
}

//fillOutItem fills OutItem item values according to attributes
func (item *Item) fillOutItemBackup(fieldName string, s *goquery.Selection) {

	if href, exists := s.Attr("href"); exists {
		m := make(map[string]interface{})
		m["href"] = href
		m["text"] = strings.TrimSpace(s.Text())
		item.value[fieldName] = m

	} else if src, exists := s.Attr("src"); exists {
		m := make(map[string]interface{})
		m["src"] = src
		m["alt"] = strings.TrimSpace(s.AttrOr("alt", ""))
		item.value[fieldName] = m
	} else {
		item.value[fieldName] = strings.TrimSpace(s.Text())
		//item[fieldName] = strings.TrimSpace(s.Text())
	}

}

func attrOrDataValue(s *goquery.Selection) (value string) {
	attr, exists := s.Attr("class")
	if exists {
		//return fmt.Sprintf("%s%s", ".", attr)
		return fmt.Sprintf(".%s", strings.Replace(strings.TrimSpace(attr), " ", ".", -1))
	}
	attr, exists = s.Attr("id")
	if exists {
		return fmt.Sprintf("#%s", attr)
	}

	return s.Nodes[0].Data
}

func dataValue(s *goquery.Selection) (value string) {
	return s.Nodes[0].Data
}

//trying to determine common parent
func (p *payload) parseItem(h []byte) (outItem outItem, err error) {
	//var pr = fmt.Println
	outItem.Name = p.Name
	outItem.URL = p.URL
	if len(p.Fields) == 0 {
		return outItem, errNoSelectors
	}
	node, err := html.Parse(bytes.NewReader(h))

	if err != nil {
		return outItem, err
	}
	doc := goquery.NewDocumentFromNode(node)
	if err != nil {
		return outItem, err
	}

	//Find closest intersection of all parents for payload fields
	parents := make(map[string]*goquery.Selection)
	var intersection *goquery.Selection
	for i, f := range p.Fields {
		parents[f.CSSSelector] = doc.Find(f.CSSSelector).Parents()
		if i == 0 {
			intersection = parents[f.CSSSelector]
		} else {
			intersection = intersection.Intersection(parents[f.CSSSelector])
		}
		//pr(intersection.Length())
		sel := doc.Find(f.CSSSelector)
		outItem.genAttrFieldName(f.FieldName, sel)
	}
	if intersection.Length() == 0 {
		return outItem, errNoSelectors
	}
	//pr(intersection.Length())
	//Adding Intersection parent to the path for more precise.
	//	intAttr := attrOrDataValue(intersection)
	//	if strings.Contains(intAttr,"#"){
	//		intAttr = dataValue(intersection)
	//	}

	intersectionWithParent := fmt.Sprintf("%s>%s",
		attrOrDataValue(intersection.Parent()),
		attrOrDataValue(intersection))
	//dataValue(intersection))
	//intersectionWithParent = attrOrDataValue(intersection)

	//pr("intParent", intersectionWithParent)
	items := doc.Find(intersectionWithParent)

	//	pr("items", items.Length())
	var inters *goquery.Selection
	if items.Length() == 1 {
		inters = items.Children()
	}
	if items.Length() > 1 {
		inters = items
	}

	inters.Each(func(i int, s *goquery.Selection) {
		//pr(i, attrOrDataValue(s))
		item := Item{value: make(map[string]interface{})}
		for _, field := range p.Fields {
			filtered := s.Find(field.CSSSelector)
			//pr(field.FieldName)
			if filtered.Length() >= 1 {
				item.fillOutItem(field.FieldName, filtered)
			}
		}
		if len(item.value) > 0 {
			outItem.Items = append(outItem.Items, item.value)

		}
	})
	outItem.Count = len(outItem.Items)
	outItem.CreatedAt = time.Now().UnixNano()
	return outItem, nil
}

//generateTable create table used by MarshalExcelSheet and MarshalCSVItem
func (o outItem) generateTable() (buf [][]string) {
	header := true
	if header {
		buf = append(buf, o.Fields)
	}
	fmt.Println(o.Fields)

	fCount := len(o.Fields)
	for _, item := range o.Items { //rows
		row := make([]string, fCount, fCount)
		var keys []string
		for i, f := range o.Fields { //field names set
			for k, v := range item.(map[string]interface{}) { //fields in row
				switch v := v.(type) {
				case map[string]interface{}:
					for k1, v1 := range v {
						joinedFieldName := fmt.Sprintf("%s_%s", k, k1)
						if joinedFieldName == f {
							row[i] = v1.(string)
							keys = append(keys, joinedFieldName)
						}
					}
				default:
					if k == f {
						row[i] = v.(string)
						keys = append(keys, k)
					}
				}
			}
		}

		for i, f := range o.Fields {
			if !stringInSlice(f, keys) {
				row[i] = ""
			}
		}
		buf = append(buf, row)
	}
	return
}

//stringInSlice check if specified string in the slice of strings
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// InsertStringToSlice inserts the value into the slice at the specified index,
// which must be in range.
// The slice must have room for the new element.
func InsertStringToSlice(slice []string, index int, value string) []string {
	// Grow the slice by one element.
	slice = slice[0 : len(slice)+1]
	// Use copy to move the upper part of the slice out of the way and open a hole.
	copy(slice[index+1:], slice[index:])
	// Store the new value.
	slice[index] = value
	// Return the result.
	return slice
}

func addStringSliceToSlice(in []string, out []string) {
	for _, s := range in {
		if !stringInSlice(s, out) {
			out = append(out, s)
		}
	}
}

//trying to determine common parent
func (p *payload) parseItem6(h []byte) (outItem outItem, err error) {
	var pr = fmt.Println
	outItem.Name = p.Name
	outItem.URL = p.URL
	if len(p.Fields) == 0 {
		return outItem, errNoSelectors
	}
	node, err := html.Parse(bytes.NewReader(h))

	if err != nil {
		return outItem, err
	}
	doc := goquery.NewDocumentFromNode(node)
	if err != nil {
		return outItem, err
	}

	parents := make(map[string]*goquery.Selection)
	var intersection *goquery.Selection
	for i, f := range p.Fields {
		parents[f.CSSSelector] = doc.Find(f.CSSSelector).Parents()
		//	pr(f.CSSSelector, parents[f.CSSSelector].Length())
		if i == 0 {
			intersection = parents[f.CSSSelector]
		} else {
			intersection = intersection.Intersection(parents[f.CSSSelector])
		}
		pr(intersection.Length())
		//pr(parents[f.CSSSelector].Length())
	}
	//intersection.Each(func(i int, s *goquery.Selection) {
	//	pr(s.Nodes[0].Data, attrOrDataValue(s))
	//})
	//pr(attrOrDataValue(intersection.First()),intersection.First().Length())
	//intersection.Each(func(i int, s *goquery.Selection) {
	//	pr(attrOrDataValue(s),s.Length())
	//})
	pr("--")
	inters := intersection
	inters.Each(func(i int, s *goquery.Selection) {
		pr("inters", attrOrDataValue(s))
	})
	pr("--")
	hasSimilarChildren := false
	children := intersection.First().Children()
	var attrs []string
	children.Each(func(i int, s *goquery.Selection) {
		attrs = append(attrs, attrOrDataValue(s))
	})
	attrStr := strings.Join(attrs, ",")
	pr("ATTRS", attrs)
	for _, attr := range attrs {
		//Looking for several similar items(list of items)
		if strings.Count(attrStr, attr) > 1 {
			hasSimilarChildren = true
			//break

		}
	}
	pr("hasSimilarChildren", hasSimilarChildren)
	pr("--")
	if !hasSimilarChildren {
		hasSimilarSibs := false
		sibs := intersection.First().Siblings()
		var siblings *goquery.Selection
		//i :=0
		//	for i := 0; i<5; i++  {
		for !hasSimilarSibs {
			var attrs []string
			siblings = sibs.Siblings()
			siblings.Each(func(i int, s *goquery.Selection) {
				attrs = append(attrs, attrOrDataValue(s))
			})
			attrStr := strings.Join(attrs, ",")
			pr("ATTRS", attrs)
			for _, attr := range attrs {
				//Looking for several similar items(list of items)
				if strings.Count(attrStr, attr) > 1 {
					hasSimilarSibs = true
					//break
				}
			}
			if !hasSimilarSibs {
				sibs = sibs.Parent()
			}
		}
		sibs.Each(func(i int, s *goquery.Selection) {
			pr("sibs", attrOrDataValue(s))
		})

		children = sibs.Parent().Children()
		children.Each(func(i int, s *goquery.Selection) {
			pr("children", attrOrDataValue(s))
		})
	}

	//nodes := doc.FindSelection(intersection.First())
	//nodes.Each(func(i int, s *goquery.Selection) {
	//	pr("nodes", attrOrDataValue(s))
	//})
	//children := intersection.First().Parent().Children()
	//children := intersection.First().AddSelection(intersection.Siblings())
	//siblings := intersection.Parent().Siblings() //najnakup
	//children := intersection.Parent().Children() //heureka
	//siblings := intersection.First().Siblings()

	//pr("sibl: ", attrOrDataValue(siblings), siblings.Length(), " child: ", attrOrDataValue(children), children.Length())

	children.Each(func(i int, s *goquery.Selection) {
		//attr, _ := s.Attr("class")
		//pr(i, attr, s.Text())
		item := Item{value: make(map[string]interface{})}
		for _, field := range p.Fields {
			filtered := s.Find(field.CSSSelector)
			if filtered.Length() >= 1 {
				item.fillOutItem(field.FieldName, filtered)
			}
		}

		if len(item.value) > 0 {
			outItem.Items = append(outItem.Items, item.value)

		}
	})
	outItem.CreatedAt = time.Now().UnixNano()
	return outItem, nil
}

//trying to determine common parent
func (p *payload) parseItemBackup(h []byte) (outItem outItem, err error) {
	var pr = fmt.Println
	outItem.Name = p.Name
	outItem.URL = p.URL
	if len(p.Fields) == 0 {
		return outItem, errNoSelectors
	}
	node, err := html.Parse(bytes.NewReader(h))

	if err != nil {
		return outItem, err
	}
	doc := goquery.NewDocumentFromNode(node)
	if err != nil {
		return outItem, err
	}

	parent := doc.Find(p.Fields[0].CSSSelector).Parent().First()
	siblings := parent.Siblings()

	//Detecting parent blocks
	found := false
	//for i:=0; i<5; i++{
	for !found {
		if siblings.Length() > 1 {
			var attrs []string
			siblings.Each(func(i int, s *goquery.Selection) {
				//if value := attrOrDataValue(s); value != "" {
				//if class, exists := s.Attr("class"); exists {
				//fmt.Println(i, "Class:", class)
				//	attrs = append(attrs, value)
				attrs = append(attrs, attrOrDataValue(s))
				//}

			})
			attrStr := strings.Join(attrs, ",")
			pr(attrs)
			for _, attr := range attrs {
				//Looking for several similar items(list of items)
				if strings.Count(attrStr, attr) > 1 {
					found = true
				}
			}

		}
		if !found {
			parent = parent.Parent().First()
			siblings = parent.Siblings()
		}
	}
	children := siblings.Parent().Children()

	children.Each(func(i int, s *goquery.Selection) {
		//pr(i, classOrIDValue(s))
		item := Item{value: make(map[string]interface{})}
		for _, field := range p.Fields {
			filtered := s.Find(field.CSSSelector)
			if filtered.Length() >= 1 {
				item.fillOutItem(field.FieldName, filtered)
			}
		}

		if len(item.value) > 0 {
			outItem.Items = append(outItem.Items, item.value)

		}
	})
	outItem.CreatedAt = time.Now().UnixNano()
	return outItem, nil
}

func (p *payload) parseItemGroupsExperiment(h []byte) (outItem outItem, err error) {
	outItem.Name = p.Name
	outItem.URL = p.URL
	if len(p.Fields) == 0 {
		return outItem, errNoSelectors
	}
	node, err := html.Parse(bytes.NewReader(h))

	if err != nil {
		return outItem, err
	}
	doc := goquery.NewDocumentFromNode(node)
	if err != nil {
		return outItem, err
	}
	var selectors []string
	for _, field := range p.Fields {
		selectors = append(selectors, field.CSSSelector)
		fmt.Println(field.CSSSelector, doc.Find(field.CSSSelector).Length())
	}
	selectorsStr := strings.Join(selectors, ",")
	items := doc.Find(selectorsStr)

	//determine the first node Selector
	var firstItem field
	for _, field := range p.Fields {
		if items.Eq(0).Is(field.CSSSelector) {
			firstItem = field
			break
		}
	}
	fmt.Println(firstItem)
	//firstItem.FieldName = "Price"
	//firstItem.CSSSelector = ".pricen"
	/*
		for i := range items.Nodes {
			curItem := items.Eq(i)
			for _, field := range p.Fields {
					if curItem.Is(field.CSSSelector) {
						fmt.Println(i, curItem.IndexSelector(field.CSSSelector), field.FieldName)
						break
					}
				}
			//fmt.Println(i)

		}
	*/

	item := Item{value: make(map[string]interface{})}
	for i := range items.Nodes {
		curItem := items.Eq(i)
		//fmt.Println(curItem.Parent().Nodes[0])
		firstSel := curItem.IndexSelector(firstItem.CSSSelector)
		//fmt.Println(i, firstSel)
		switch {
		case firstSel == 0:
			item.fillOutItem(firstItem.FieldName, curItem)
			fmt.Println(i, firstSel, firstItem.FieldName)
		case firstSel < 0:
			for _, field := range p.Fields {
				if curItem.Is(field.CSSSelector) {
					item.fillOutItem(field.FieldName, curItem)
					fmt.Println(i, firstSel, field.FieldName)
					break
				}
			}

		case firstSel > 0:
			outItem.Items = append(outItem.Items, item.value)
			item = Item{value: make(map[string]interface{})}
			item.fillOutItem(firstItem.FieldName, curItem)
			fmt.Println(i, firstSel, firstItem.FieldName)
		}
		//adding last item to output structure
		if i == items.Length()-1 {
			outItem.Items = append(outItem.Items, item.value)
		}

	}
	outItem.CreatedAt = time.Now().UnixNano()
	return outItem, nil
}

//trying to determine common parent
func (p *payload) parseItem2(h []byte) (outItem outItem, err error) {
	var pr = fmt.Println
	d, err := goquery.NewDocument("http://drony.heureka.sk")
	if err != nil {
		pr(err)
	}
	pr(d)
	pr("___")
	outItem.Name = p.Name
	outItem.URL = p.URL
	if len(p.Fields) == 0 {
		return outItem, errNoSelectors
	}
	node, err := html.Parse(bytes.NewReader(h))

	if err != nil {
		return outItem, err
	}
	doc := goquery.NewDocumentFromNode(node)
	if err != nil {
		return outItem, err
	}

	parent := doc.Find(p.Fields[0].CSSSelector).Parent().First()
	//attr, _ := parent.Attr("class")
	//pr(parent.Attr("class"))
	siblings := parent.Siblings()
	pr("siblings len", parent.Siblings().Length())
	//pr(siblings.Length())
	//siblings.Each(func(i int, s *goquery.Selection) {
	//	pr(s.Nodes[0].Data)
	//})

	//Detecting parent blocks
	found := false
	//for i := 0; i < 5; i++ {
	for !found {
		//for parent.Parent().Siblings().Length() != 0 {
		similarClassesFound := false
		if siblings.Length() > 1 {

			var classes []string
			//var data []string
			siblings.Each(func(i int, s *goquery.Selection) {
				if class, exists := s.Attr("class"); exists {
					//fmt.Println(i, "Class:", class)
					classes = append(classes, class)
				}
				//data = append(data, s.Nodes[0].Data)

			})
			classesStr := strings.Join(classes, ",")
			classesMap := make(map[string]int)
			for _, class := range classes {
				//Looking for several similar items(list of items)
				classesMap[class] = strings.Count(classesStr, class)
				//	if strings.Count(classesStr, class) > 1 {
				//		found = true
				//	}
			}
			//fmt.Println("Classes", classesMap)
			for k, v := range classesMap {
				if v > 1 {

					//if parent.Parent().Siblings().Length() == 0 {
					//	if true{
					similarClassesFound = true
					//	}
				}
				fmt.Println("Classes", k, v, similarClassesFound)
			}
			//fmt.Println("Classes", classes)
			//fmt.Println("Data", data)
		}
		//if parent.Parent().Siblings().Length() != 0 {
		//
		//	found = false
		//}

		//if !found {
		//if parent.Parent().Siblings().Length() != 0 {
		id, exists := parent.Parent().Attr("id")
		pr("ID", id, exists)
		found = exists
		if similarClassesFound {
			class, exists := parent.Parent().Attr("class")
			pr("URRRAAA", class, exists)
			if exists && parent.Parent().Siblings().Length() > 0 {
				pr("URRRAAA2", class, exists)
				parent.Parent().Siblings().Each(func(i int, s *goquery.Selection) {
					if s.HasClass(class) {
						found = false
						pr("URRRAAA3", class, exists)

					}
				})
			}
			//	if parent.Parent().Siblings().Length() == 0 { //esli net sosedei s tem je klassom
			//		pr("FOUNDSSAAA")
			//		found = true
			//	}
		}

		if !found {
			parent = parent.Parent().First()
			pr("siblings len", parent.Siblings().Length())
			//pr("TRRRR", parent.Parent().Siblings().Length())
			siblings = parent.Siblings()
		}
	}

	pr(parent.Parent().Attr("class"))
	pr(parent.Parent().Length())
	children := siblings.Parent().Children()
	//pr(siblings.Parent().Nodes[0])

	children.Each(func(i int, s *goquery.Selection) {
		//attr, _ := s.Attr("class")
		//pr(i, attr, s.Text())
		item := Item{value: make(map[string]interface{})}
		for _, field := range p.Fields {
			filtered := s.Find(field.CSSSelector)
			if filtered.Length() >= 1 {
				item.fillOutItem(field.FieldName, filtered)
			}
		}

		if len(item.value) > 0 {
			outItem.Items = append(outItem.Items, item.value)

		}
	})
	outItem.CreatedAt = time.Now().UnixNano()
	return outItem, nil
}

//trying to determine common parent
func (p *payload) parseItem4(h []byte) (outItem outItem, err error) {
	var pr = fmt.Println
	outItem.Name = p.Name
	outItem.URL = p.URL
	if len(p.Fields) == 0 {
		return outItem, errNoSelectors
	}
	node, err := html.Parse(bytes.NewReader(h))

	if err != nil {
		return outItem, err
	}
	doc := goquery.NewDocumentFromNode(node)
	if err != nil {
		return outItem, err
	}

	parent := doc.Find(p.Fields[0].CSSSelector).Parent().First()
	siblings := parent.Siblings()

	//Detecting parent blocks
	found := false
	for !found {
		similarClassesFound := false
		if siblings.Length() > 1 {
			var classes []string
			var data []string
			siblings.Each(func(i int, s *goquery.Selection) {
				if class, exists := s.Attr("class"); exists {
					//fmt.Println(i, "Class:", class)
					classes = append(classes, class)
				}
				data = append(data, s.Nodes[0].Data)
			})
			classesStr := strings.Join(classes, ",")
			for _, class := range classes {
				//Looking for several similar items(list of items)
				if strings.Count(classesStr, class) > 1 {
					similarClassesFound = true
				}
			}
			fmt.Println("Classes", classes, data)
		}
		if !found && !similarClassesFound {
			parent = parent.Parent().First()
			siblings = parent.Siblings()
		}
	}
	children := siblings.Parent().Children()
	pr("_________")
	children.Each(func(i int, s *goquery.Selection) {
		attr, _ := s.Attr("class")
		//pr(i, attr, s.Text())
		pr(i, attr, s.Nodes[0].Data)

		item := Item{value: make(map[string]interface{})}
		for _, field := range p.Fields {
			filtered := s.Find(field.CSSSelector)
			if filtered.Length() >= 1 {
				item.fillOutItem(field.FieldName, filtered)
			}
		}

		if len(item.value) > 0 {
			outItem.Items = append(outItem.Items, item.value)

		}
	})
	outItem.CreatedAt = time.Now().UnixNano()
	return outItem, nil
}
