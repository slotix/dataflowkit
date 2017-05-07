package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/slotix/dataflowkit/helpers"
	"github.com/slotix/dataflowkit/splash"
	"golang.org/x/net/html"
	"gopkg.in/go-playground/validator.v9"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "parser: ", log.Lshortfile)
	validate = validator.New()
}

var errNoSelectors = errors.New("No selectors found")
var errEmptyURL = errors.New("no URL provided")

//NewParser initializes new Parser struct
func NewParser(payload []byte) (Parser, error) {
	var p Parser
	err := p.UnmarshalJSON(payload)
	if err != nil {
		return p, err
	}
	if p.Format == "" {
		p.Format = "json"
	}
	p.PayloadMD5 = helpers.GenerateMD5(payload)
	return p, nil
}

//Parse parses payload json structure and generate Out to be serializes as JSON, XML, CSV, Excel
func (p *Parser) Parse() (*Collections, error) {
	//parse input and fill Payload structure
	out := Collections{}
	for _, collection := range p.Payloads {
		req := splash.Request{URL: collection.URL}
		splashURL, err := splash.NewSplashConn(req)
		content, err := splash.Fetch(splashURL)
		if err != nil {
			logger.Println(err)
			//return nil, err
		}
		outItem, err := collection.parseItem(content)
		if err != nil {
			logger.Println(err)
			//return out, err
			//logger.Printf("\"%s:\" %s\n", outItem.Name, err)
		} else {
			out.Collections = append(out.Collections, outItem)
		}
	}

	if len(out.Collections) == 0 {
		return nil, errNoSelectors
	}
	return &out, nil
}

//NewCollection initializes new collection
func newCollection(p *payload) (*collection, error) {
	meta := meta{
		Name: p.Name,
		URL:  p.URL,
	}
	//logger.Println(meta)
	err := validate.Struct(meta)
	if err != nil {
		return nil, err
	}
	c := collection{
		meta: meta,
	}
	//if p.URL == "" {
	//	return &c, errEmptyURL
	//}
	if len(p.Fields) == 0 {
		return nil, errNoSelectors
	}
	return &c, nil
}

//trying to determine common parent
func (p *payload) parseItem1(h []byte) (col *collection, err error) {
	col, err = newCollection(p)
	if err != nil {
		return nil, err
	}
	node, err := html.Parse(bytes.NewReader(h))

	if err != nil {
		return nil, err
	}
	doc := goquery.NewDocumentFromNode(node)
	if err != nil {
		return nil, err
	}

	//findSelectors(doc, "a")

	//Find closest intersection of all parents for payload fields
	parents := make(map[string]*goquery.Selection)
	var intersection *goquery.Selection
	for i, f := range p.Fields {
		err := validate.Struct(f)
		if err != nil {
			return nil, err
		}
		sel := doc.Find(f.CSSSelector)
		logger.Println(f.CSSSelector, sel.Length())
		col.genAttrFieldName(f.Name, sel)
		parents[f.CSSSelector] = doc.Find(f.CSSSelector).Parents()
		if sel.Length() > 0 { //don't add selectors to intersection if lenght is 0. Otherwise the whole intersection returns No selectors error
			if i == 0 {
				intersection = parents[f.CSSSelector]
			} else {
				intersection = intersection.Intersection(parents[f.CSSSelector])
			}

		}
	}
	logger.Println(attrOrDataValue(intersection))
	if intersection == nil || intersection.Length() == 0 {
		return nil, errNoSelectors
	}

	intersectionWithParent := fmt.Sprintf("%s>%s",
		attrOrDataValue(intersection.Parent()),
		attrOrDataValue(intersection))
	logger.Println(intersectionWithParent)

	items := doc.Find(intersectionWithParent)
	logger.Println(items.Length())
	var inter1 *goquery.Selection
	if items.Length() == 1 {
		inter1 = items.Children()
	}
	if items.Length() > 1 {
		inter1 = items
	}

	inter1.Each(func(i int, s *goquery.Selection) {
		logger.Println(i, attrOrDataValue(s))
		itm := item{value: make(map[string]interface{})}
		for _, field := range p.Fields {
			filtered := s.Find(field.CSSSelector)
			//pr(field.FieldName)
			if filtered.Length() >= 1 {
				itm.fillCollection(field.Name, filtered)
			}
		}
		if len(itm.value) > 0 {
			col.Items = append(col.Items, itm.value)

		}
	})
	col.Count = len(col.Items)
	col.CreatedAt = time.Now().UnixNano()
	return col, nil
}
func intersectionFL(sel *goquery.Selection) *goquery.Selection {
	first := sel.First()
	last := sel.Last()
	intersection := first.Parents().Intersection(last.Parents())
	return intersection
}

func (p *payload) parseItem(r io.Reader) (col *collection, err error) {
	col, err = newCollection(p)
	if err != nil {
		return nil, err
	}
	//node, err := html.Parse(bytes.NewReader(h))

	//if err != nil {
	//	return nil, err
	//}
	//	doc := goquery.NewDocumentFromNode(node)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	var intersection *goquery.Selection
	for i, f := range p.Fields {
		err := validate.Struct(f)
		if err != nil {
			return nil, err
		}
		sel := doc.Find(f.CSSSelector)
		logger.Println(f.CSSSelector, sel.Length())
		col.genAttrFieldName(f.Name, sel)
		if sel.Length() > 0 { //don't add selectors to intersection if lenght is 0. Otherwise the whole intersection returns No selectors error
			if i == 0 {
				intersection = intersectionFL(sel)
			} else {
				intersection = intersection.Intersection(intersectionFL(sel))
			}
		}
	}
	if intersection == nil || intersection.Length() == 0 {
		return nil, errNoSelectors
	}
	intersectionWithParent := fmt.Sprintf("%s>%s",
		attrOrDataValue(intersection.Parent()),
		attrOrDataValue(intersection))
	logger.Println(intersectionWithParent)
	items := doc.Find(intersectionWithParent)
	//logger.Println(items.Length())
	var inter1 *goquery.Selection
	if items.Length() == 1 {
		inter1 = items.Children()
	}
	if items.Length() > 1 {
		inter1 = items
	}
	var itms []item
	inter1.Each(func(i int, s *goquery.Selection) {
		//logger.Println(i, attrOrDataValue(s))
		itm := item{value: make(map[string]interface{})}
		for _, field := range p.Fields {
			filtered := s.Find(field.CSSSelector)
			if filtered.Length() > 1 {
				filtered.Each(func(i int, s *goquery.Selection) {
					itm1 := item{value: make(map[string]interface{})}
					itm1.fillCollection(field.Name, s)
					itms = append(itms, itm1)
				})

			} else if filtered.Length() == 1 {
				itm.fillCollection(field.Name, filtered)
			}
		}

		//	logger.Println(itms)
		if len(itms) > 0 {
			for _, i := range itms {
				col.Items = append(col.Items, i.value)
			}
		} else if len(itm.value) > 0 {
			col.Items = append(col.Items, itm.value)
		}
	})

	/*
		inter1.Each(func(i int, s *goquery.Selection) {
			itm := item{value: make(map[string]interface{})}
			//var itms []item
			for _, field := range p.Fields {
				filtered := s.Find(field.CSSSelector)
				if filtered.Length() >= 1 {
					filtered.Each(func(i int, s *goquery.Selection) {
						itm.fillCollection(field.Name, s)
					//	itms = append(itms, itm)
					})
				}
				if filtered.Length() >= 1 {
					itm.fillCollection(field.Name, filtered)
				}
			}
			//logger.Println(len(itms))

			//if len(itms) > 0 {
			//	for _, i := range itms {
			//		col.Items = append(col.Items, i.value)
			//	}
			//}
			//if len(itm.value) > 0 {
			//	col.Items = append(col.Items, itm.value)

			//}
		})*/
	col.Count = len(col.Items)
	col.CreatedAt = time.Now().UnixNano()
	return col, nil
}

//generateTable create table used by MarshalExcelSheet and MarshalCSVItem
func (c collection) generateTable() (buf [][]string) {
	header := true
	if header {
		buf = append(buf, c.Fields)
	}
	logger.Println(c.Fields)

	fCount := len(c.Fields)
	for _, item := range c.Items { //rows
		row := make([]string, fCount, fCount)
		var keys []string
		for i, f := range c.Fields { //field names set
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

		for i, f := range c.Fields {
			if !helpers.StringInSlice(f, keys) {
				row[i] = ""
			}
		}
		buf = append(buf, row)
	}
	return
}

func findSelectors(doc *goquery.Document, what string) {
	sel := doc.Find(what)
	sel.Each(func(i int, s *goquery.Selection) {
		attr, _ := s.Attr("href")
		logger.Println(i, attr)
	})
}

//MarshalData parses payload raw JSON data and generates output
//Here is an example of payload structure:
/*
{"format":"json",
	"collections": [
            {
            "name": "collection1",
            "url": "http://example1.com",
            "fields": [
                {
                    "field_name": "link",
                    "css_selector": ".link a"
                },
                {
                    "field_name": "Text",
                    "css_selector": ".text"
                },
				{
					"field_name": "Image",
					"css_selector": ".foto img"
				}
            ]
        }
    ]
}
*/
//func MarshalData(payload []byte) ([]byte, error) {
func (p *Parser) MarshalData() (io.ReadCloser, error) {
	cols, err := p.Parse()
	if err != nil {
		return nil, err
	}
	var b []byte
	switch p.Format {
	case "xml":
		b, err = cols.MarshalXML()
	case "csv":
		b, err = cols.MarshalCSV("")
	default:
		b, err = cols.MarshalJSON()
	}
	if err != nil {
		return nil, err
	}
	readCloser := ioutil.NopCloser(bytes.NewReader(b))
	return readCloser, nil
}

//genAttrFieldName generates field name according to attributes
func (o *collection) genAttrFieldName(fieldName string, sel *goquery.Selection) {
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

type item struct {
	value map[string]interface{}
}

//fillCollection fills Collection item values according to attributes
func (i *item) fillCollection(fieldName string, s *goquery.Selection) {

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
				i.value[fieldName] = m
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
				i.value[fieldName] = m
			}
		} else {
			i.value[fieldName] = strings.TrimSpace(s.Text())
		}
	}
}

func attrOrDataValue(s *goquery.Selection) (value string) {
	attr, exists := s.Attr("class")
	if exists && attr != "" { //in some cases tag is invalid f.e. <tr class>
		return fmt.Sprintf(".%s", strings.Replace(strings.TrimSpace(attr), " ", ".", -1))
	}
	attr, exists = s.Attr("id")

	if exists && attr != "" {
		return fmt.Sprintf("#%s", attr)
	}
	return s.Nodes[0].Data
}

func dataValue(s *goquery.Selection) (value string) {
	return s.Nodes[0].Data
}
