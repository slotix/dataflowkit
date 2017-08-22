package scrape

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/helpers"
	"github.com/slotix/dataflowkit/paginate"
)

//NewParser initializes new Parser struct
func NewPayload(payload []byte) (Payload, error) {
	var p Payload
	err := json.Unmarshal(payload, &p)
	//err := p.UnmarshalJSON(payload)
	if err != nil {
		return p, err
	}
	if p.Format == "" {
		p.Format = DefaultOptions.Format
	}
	if p.RetryTimes == 0 {
		p.RetryTimes = DefaultOptions.RetryTimes
	}
	if p.FetchDelay == 0 {
		p.FetchDelay = DefaultOptions.FetchDelay
	}
	if p.RandomizeFetchDelay == nil {
		p.RandomizeFetchDelay = &DefaultOptions.RandomizeFetchDelay
	}
	if p.PaginateResults == nil {
		p.PaginateResults = &DefaultOptions.PaginateResults
	}

	p.PayloadMD5 = helpers.GenerateMD5(payload)
	return p, nil
}

func (p Payload) PayloadToScrapeConfig() (config *ScrapeConfig, err error) {
	fetcher, err := NewSplashFetcher()
	if err != nil {
		logger.Println(err)
	}
	pieces := []Piece{}
	selectors := []string{}
	names := []string{}
	for _, f := range p.Fields {
		//var extractor scrape.PieceExtractor
		params := make(map[string]interface{})
		if f.Extractor.Params != nil {
			params = f.Extractor.Params.(map[string]interface{})
		}
		switch eType := f.Extractor.Type; eType {
		//For Link type by default Two pieces with different Text and Attr="href" extractors will be added for field selector.
		case "link":
			t := &extract.Text{}
			if params != nil {
				err := extract.FillStruct(params, t)
				if err != nil {
					logger.Println(err)
				}
			}
			fName := fmt.Sprintf("%s_text", f.Name)
			pieces = append(pieces, Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: t,
			})
			names = append(names, fName)

			a := &extract.Attr{Attr: "href"}
			if params != nil {
				err := extract.FillStruct(params, a)
				if err != nil {
					logger.Println(err)
				}
			}
			fName = fmt.Sprintf("%s_link", f.Name)
			pieces = append(pieces, Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: a,
			})
			names = append(names, fName)
			//Add selector just one time for link type
			selectors = append(selectors, f.Selector)
		//For image type by default Two pieces with different Attr="src" and Attr="alt" extractors will be added for field selector.
		case "image":
			a := &extract.Attr{Attr: "src"}
			if params != nil {
				err := extract.FillStruct(params, a)
				if err != nil {
					logger.Println(err)
				}
			}
			fName := fmt.Sprintf("%s_src", f.Name)
			pieces = append(pieces, Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: a,
			})
			names = append(names, fName)

			a = &extract.Attr{Attr: "alt"}
			if params != nil {
				err := extract.FillStruct(params, a)
				if err != nil {
					logger.Println(err)
				}
			}
			fName = fmt.Sprintf("%s_alt", f.Name)
			pieces = append(pieces, Piece{
				Name:      fName,
				Selector:  f.Selector,
				Extractor: a,
			})
			names = append(names, fName)
			//Add selector just one time for link type
			selectors = append(selectors, f.Selector)

		default:
			var e extract.PieceExtractor
			switch eType {
			case "const":
				c := &extract.Const{Val: params["value"]}
				e = c
			case "count":
				e = &extract.Count{}
			case "text":
				e = &extract.Text{}
			case "multipleText":
				e = &extract.MultipleText{}
			case "html":
				e = &extract.Html{}
			case "outerHtml":
				e = &extract.OuterHtml{}
			case "attr":
				e = &extract.Attr{}
			case "regex":
				r := &extract.Regex{}
				regExp := params["regexp"]
				r.Regex = regexp.MustCompile(regExp.(string))
				e = r
			}
			//extractor, err := extract.FillParams(f.Extractor.Type, params)
			//err := e.FillParams(params)
			if params != nil {
				err := extract.FillStruct(params, e)
				if err != nil {
					logger.Println(err)
				}
			}
			pieces = append(pieces, Piece{
				Name:      f.Name,
				Selector:  f.Selector,
				Extractor: e,
			})
			selectors = append(selectors, f.Selector)
			names = append(names, f.Name)
		}
	}

	paginator := p.Paginator
	config = &ScrapeConfig{
		Fetcher: fetcher,
		//DividePage: scrape.DividePageBySelector(".p"),
		DividePage: DividePageByIntersection(selectors),
		Pieces:     pieces,
		//Paginator: paginate.BySelector(".next", "href"),
		CSVHeader: names,
		Paginator: paginate.BySelector(paginator.Selector, paginator.Attribute),
		Opts: ScrapeOptions{
			MaxPages:            paginator.MaxPages,
			Format:              p.Format,
			PaginateResults:     *p.PaginateResults,
			FetchDelay:          p.FetchDelay,
			RandomizeFetchDelay: *p.RandomizeFetchDelay,
			RetryTimes:          p.RetryTimes,
		},
	}
	return
}
