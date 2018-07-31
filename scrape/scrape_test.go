package scrape

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	randomize               bool
	delayFetch              time.Duration
	paginateResults         bool
	pJSON, pCSV_XML         Payload
	outJSON, outCSV, outXML []byte
)

func init() {
	viper.Set("CHROME", "http://127.0.0.1:9222")
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	viper.Set("STORAGE_TYPE", "diskv")
	viper.Set("RESULTS_DIR", "results")
	viper.Set("RANDOMIZE_FETCH_DELAY", true)
	randomize = true
	//delayFetch = 500 * time.Millisecond
	delayFetch = 0
	paginateResults = false
}

func TestScraper_partNames(t *testing.T) {
	s := Scraper{}
	s.Parts = []Part{
		Part{Name: "1"},
		Part{Name: "2"},
		Part{Name: "3"},
		Part{Name: "4"},
	}
	parts := s.partNames()
	assert.Equal(t, []string{"1", "2", "3", "4"}, parts)

}

func TestPayload_selectors(t *testing.T) {
	p1 := Payload{
		Fields: []Field{
			Field{Selector: "sel1"},
			Field{Selector: "sel2"},
			Field{Selector: "sel3"},
			Field{Selector: "sel4"},
		},
	}
	p2 := Payload{
		Fields: []Field{
			Field{},
			Field{},
			Field{},
			Field{},
		},
	}

	s1, err := p1.selectors()
	assert.NoError(t, err)
	assert.Equal(t, []string{"sel1", "sel2", "sel3", "sel4"}, s1)
	s2, err := p2.selectors()
	assert.Error(t, err)
	assert.Equal(t, []string(nil), s2)

}

func TestNewTask(t *testing.T) {
	task := NewTask(Payload{})
	assert.NotEmpty(t, task.ID)
	start, err := task.startTime()
	assert.NoError(t, err)
	assert.NotNil(t, start, "task start time is not nil")
	//t.Log(start)
}

var update = flag.Bool("update", false, "update .golden files")

func TestParse(t *testing.T) {
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	paginateResults = false
	p := Payload{
		Name: "persons Table",
		Request: fetch.Request{
			Type: "chrome",
			URL:  "http://testserver:12345/persons/page-0",
		},
		Fields: []Field{
			Field{
				Name:     "Names",
				Selector: "#cards a",
				Extractor: Extractor{
					Types: []string{"text"}, // "const", "outerHtml"},
					// Params: map[string]interface{}{
					// 	"value": "2",
					// },
				},
			},
			Field{
				Name:     "Images",
				Selector: ".card-img-top",
				Extractor: Extractor{
					Types: []string{"src", "alt"},
				},
			},
			// Field{
			// 	Name:     "Count",
			// 	Selector: "#cards a",
			// 	Extractor: Extractor{
			// 		Types: []string{"count"},
			// 	},
			// },
		},
		PaginateResults: &paginateResults,
		Format:          "json",
	}

	task := NewTask(p)
	r, err := task.Parse()
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	resultFile := buf.Bytes()
	actual, err := ioutil.ReadFile(filepath.Join("./", string(resultFile)))
	assert.NoError(t, err)
	//t.Log(string(got))
	golden := filepath.Join("../testdata", "jsonRes.golden")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)

	// if !bytes.Equal(actual, expected) {
	// 	// FAIL!
	// }
	//expected := []byte(`[{"Count_count":10,"Header_const":"1","Header_outerHtml":"\u003ch1\u003ePersons\u003c/h1\u003e","Warning_html":"\u003cstrong\u003eWarning!\u003c/strong\u003e This is a demo website for web scraping purposes. \u003cbr/\u003eThe data on this page has been randomly generated."}]` + "\n")
	assert.Equal(t, expected, actual)
	///// No selectors
	badP := Payload{
		Name: "No Selectors",
		Request: fetch.Request{
			URL: "http://127.0.0.1:12345",
		},
		PaginateResults: &paginateResults,
		Format:          "json",
	}

	task = NewTask(badP)
	r, err = task.Parse()
	assert.Error(t, err, "400: no parts found")
	//Bad output format
	// badOF := Payload{
	// 	Name: "No Selectors",
	// 	Request: fetch.Request{
	// 		URL: "http://127.0.0.1:12345",
	// 	},
	// 	Fields: []Field{
	// 		Field{
	// 			Name:     "P",
	// 			Selector: "p",
	// 			Extractor: Extractor{
	// 				Types: []string{"text"},
	// 			},
	// 		},
	// 	},
	// 	PaginateResults: &paginateResults,
	// 	Format:          "BadOutputFormat",
	// }
	// task = NewTask(badOF)

	// r, err = task.Parse()
	// assert.Error(t, err, "invalid output format specified")
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
}

func TestParseSwitchFetchers(t *testing.T) {
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	// viper.Set("SKIP_STORAGE_MW", true)
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	paginateResults = false
	// p := Payload{
	// 	Name: "quotes",
	// 	Request: fetch.Request{
	// 		URL: "http://quotes.toscrape.com/js/",
	// 	},
	// 	Fields: []Field{
	// 		Field{
	// 			Name:     "quotes",
	// 			Selector: ".text",
	// 			Extractor: Extractor{
	// 				Types: []string{"text"},
	// 			},
	// 		},
	// 	},
	// 	PaginateResults: &paginateResults,
	// 	Format:          "json",
	// }
	p := Payload{
		Name: "persons Table",
		Request: fetch.Request{
			Type: "base",
			URL:  "http://testserver:12345/persons/page-0",
		},
		Fields: []Field{
			Field{
				Name:     "Names",
				Selector: "#cards a",
				Extractor: Extractor{
					Types: []string{"text"},
				},
			},
		},
		PaginateResults: &paginateResults,
		Format:          "csv",
	}
	task := NewTask(p)
	r, err := task.Parse()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
}
