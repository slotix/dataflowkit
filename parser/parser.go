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
