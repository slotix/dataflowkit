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
	//	randTrue   = true
	//	randFalse  = false
	delayFetch time.Duration
	//paginateResults                bool
	personsPayload, detailsPayload, CSVPayload, XMLPayload Payload
	update                                                 = flag.Bool("update", false, "update result files")
)

func init() {
	viper.Set("CHROME", "http://127.0.0.1:9222")
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	viper.Set("STORAGE_TYPE", "diskv")
	viper.Set("RESULTS_DIR", "results")
	viper.Set("RANDOMIZE_FETCH_DELAY", true)
	//delayFetch = 500 * time.Millisecond
	delayFetch = 0
	//paginateResults = false
	personsPayload = Payload{
		Name: "persons Cards",
		Request: fetch.Request{
			Type: "chrome",
			URL:  "http://testserver:12345/persons/page-0",
		},
		Fields: []Field{
			{
				Name:     "Names",
				Selector: "#cards a",
				Extractor: Extractor{
					Types: []string{"text", "href", "const", "outerHtml", "unknownSelectorType"},
					Params: map[string]interface{}{
						"value": "--- NAME ---",
					},
				},
			},
			{
				Name:     "Images",
				Selector: ".card-img-top",
				Extractor: Extractor{
					Types: []string{"src", "alt", "width", "height"},
				},
			},
		},
		Paginator: &paginator{
			Selector:  "nav:nth-child(4) :nth-child(2) .page-link",
			Attribute: "href",
			MaxPages:  2,
		},
		//	PaginateResults: &paginateResults,
		//RandomizeFetchDelay: &randFalse,
		Format: "json",
	}
	detailsPayload = Payload{
		Name: "persons details",
		Request: fetch.Request{
			Type: "chrome",
			URL:  "http://testserver:12345/persons/page-0",
		},
		Fields: []Field{
			{
				Name:     "Links",
				Selector: "#cards a",
				Extractor: Extractor{
					Types: []string{"path"},
					//Filters: []string{"trim"},
				},
				Details: &details{
					Fields: []Field{
						{
							Name:     "Number",
							Selector: ".display-4",
							Extractor: Extractor{
								Types: []string{"regex"},
								Params: map[string]interface{}{
									"regexp": "([\\d]+)\\s",
								},
								Filters: []string{"trim"},
							},
						},
						{
							Name:     "Name",
							Selector: ".display-4",
							Extractor: Extractor{
								Types:   []string{"text"},
								Filters: []string{"trim"},
							},
						},
						{
							Name:     "Company",
							Selector: ".card-text:nth-child(3) .col-5",
							Extractor: Extractor{
								Types:   []string{"text"},
								Filters: []string{"trim"},
							},
						},
						{
							Name:     "Phones",
							Selector: ".col-10 span",
							Extractor: Extractor{
								// Types: []string{"regex"},
								// Params: map[string]interface{}{
								// 	"regexp": "([\\d\\.]+)",
								// },
								Types:   []string{"text"},
								Filters: []string{"trim"},
							},
						},
						{
							Name:     "Email",
							Selector: ".card-text:nth-child(2) .col-5",
							Extractor: Extractor{
								Types:   []string{"text"},
								Filters: []string{"trim"},
							},
						},
					},
				},
			},
			{
				Name:     "Count",
				Selector: ".badge-primary",
				Extractor: Extractor{
					Types: []string{"count"},
				},
			},
		},
		// Paginator: &paginator{
		// 	Selector:  "nav:nth-child(4) :nth-child(2) .page-link",
		// 	Attribute: "href",
		// 	MaxPages:  2,
		// },
		//RandomizeFetchDelay: &randTrue,
		//	FetchDelay:          &delayFetch,
		Format: "json",
		//PaginateResults: &paginateResults,
	}
	CSVPayload = Payload{
		Name: "persons details",
		Request: fetch.Request{
			Type: "base",
			URL:  "http://127.0.0.1:12345/persons/3",
		},
		Fields: []Field{
			{
				Name:     "Name",
				Selector: ".display-4",
				Extractor: Extractor{
					Types:   []string{"text"},
					Filters: []string{"trim"},
				},
			},
			{
				Name:     "Phones",
				Selector: ".col-10 span",
				Extractor: Extractor{
					Types:   []string{"text"},
					Filters: []string{"trim"},
				},
			},
			{
				Name:     "PhoneCount",
				Selector: ".col-10 span",
				Extractor: Extractor{
					Types: []string{"count"},
				},
			},
			{
				Name:     "Const",
				Selector: ".col-10 span",
				Extractor: Extractor{
					Types: []string{"const", "unknownSelectorType"},
					Params: map[string]interface{}{
						"value": "--- CONST ---",
					},
				},
			},
		},
		Format: "csv",
	}
	XMLPayload = Payload{
		Name: "persons details",
		Request: fetch.Request{
			Type: "base",
			URL:  "http://127.0.0.1:12345/persons/3",
		},
		Fields: []Field{
			{
				Name:     "Name",
				Selector: ".display-4",
				Extractor: Extractor{
					Types:   []string{"text"},
					Filters: []string{"trim"},
				},
			},
			{
				Name:     "Phones",
				Selector: ".col-10 span",
				Extractor: Extractor{
					Types:   []string{"text"},
					Filters: []string{"trim"},
				},
			},
		},
		Format: "xml",
	}
}

func TestNewTask(t *testing.T) {
	viper.Set("MAX_PAGES", 10)
	task := NewTask(Payload{
		Paginator: &paginator{
			Selector:       ".paginatorrr",
			Attribute:      "href",
			InfiniteScroll: true,
		},
	})
	assert.NotEmpty(t, task.ID)
	start, err := task.startTime()
	assert.NoError(t, err)
	assert.NotNil(t, start, "task start time is not nil")
}

func TestParseDetails(t *testing.T) {
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
	viper.Set("RANDOMIZE_FETCH_DELAY", true)
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	//JSON details output
	task := NewTask(detailsPayload)
	r, err := task.Parse()
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	resultFile := buf.Bytes()
	actual, err := ioutil.ReadFile(filepath.Join("./", string(resultFile)))
	assert.NoError(t, err)
	golden := filepath.Join("../testdata", "details.json")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	//todo: test details with xml encoder
	//XML details output
	// detailsPayload.Format = "xml"
	// task = NewTask(detailsPayload)
	// r, err = task.Parse()
	// assert.NoError(t, err)
	// buf = new(bytes.Buffer)
	// buf.ReadFrom(r)
	// resultFile = buf.Bytes()
	// actual, err = ioutil.ReadFile(filepath.Join("./", string(resultFile)))
	// assert.NoError(t, err)
	// golden = filepath.Join("../testdata", "details.xml")
	// if *update {
	// 	ioutil.WriteFile(golden, actual, 0644)
	// }
	//expected, err = ioutil.ReadFile(golden)
	//assert.NoError(t, err)
	//assert.Equal(t, expected, actual)

	//todo: test details with csv encoder
	//CSV details output
	// detailsPayload.Format = "csv"
	// task = NewTask(detailsPayload)
	// r, err = task.Parse()
	// assert.NoError(t, err)
	// buf = new(bytes.Buffer)
	// buf.ReadFrom(r)
	// resultFile = buf.Bytes()
	// actual, err = ioutil.ReadFile(filepath.Join("./", string(resultFile)))
	// assert.NoError(t, err)
	// golden = filepath.Join("../testdata", "details.csv")
	// if *update {
	// 	ioutil.WriteFile(golden, actual, 0644)
	// }
	// expected, err = ioutil.ReadFile(golden)
	// assert.NoError(t, err)
	// assert.Equal(t, expected, actual)

	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
}

func TestParse(t *testing.T) {
	viper.Set("RANDOMIZE_FETCH_DELAY", false)
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	//JSON output
	task := NewTask(personsPayload)
	r, err := task.Parse()
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	resultFile := buf.Bytes()
	actual, err := ioutil.ReadFile(filepath.Join("./", string(resultFile)))
	assert.NoError(t, err)
	//t.Log(string(got))
	golden := filepath.Join("../testdata", "res.json")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	//CSV
	// personsPayload.Format = "csv"
	// task = NewTask(personsPayload)
	// r, err = task.Parse()
	// assert.NoError(t, err)
	// buf = new(bytes.Buffer)
	// buf.ReadFrom(r)
	// resultFile = buf.Bytes()
	// actual, err = ioutil.ReadFile(filepath.Join("./", string(resultFile)))
	// assert.NoError(t, err)
	// golden = filepath.Join("../testdata", "res.csv")
	// if *update {
	// 	ioutil.WriteFile(golden, actual, 0644)
	// }
	// expected, err = ioutil.ReadFile(golden)
	// assert.NoError(t, err)
	// assert.Equal(t, expected, actual)

	//xml
	personsPayload.Format = "xml"
	task = NewTask(personsPayload)
	r, err = task.Parse()
	assert.NoError(t, err)
	buf = new(bytes.Buffer)
	buf.ReadFrom(r)
	resultFile = buf.Bytes()
	actual, err = ioutil.ReadFile(filepath.Join("./", string(resultFile)))
	assert.NoError(t, err)
	golden = filepath.Join("../testdata", "res.xml")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	_, err = ioutil.ReadFile(golden)
	assert.NoError(t, err)
	//todo: order of fields in both xml files is not identical. So it is not possible to compare them easily.
	//assert.Equal(t, expected, actual)

	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
}
func TestParseErrs(t *testing.T) {
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	///// No selectors
	badP := Payload{
		Name: "No Selectors",
		Request: fetch.Request{
			URL: "http://127.0.0.1:12345",
		},
		Format: "json",
	}

	task := NewTask(badP)
	_, err := task.Parse()
	assert.Error(t, err, "400: no parts found")

	///// ErrNoPartOrSelectorProvided
	badP = Payload{
		Name: "ErrNoPartOrSelectorProvided",
		Request: fetch.Request{
			URL: "http://127.0.0.1:12345",
		},
		Fields: []Field{
			{
				Name:     "Alert",
				Selector: "",
				Extractor: Extractor{
					Types: []string{"text"},
				},
			},
			{
				Name:     "",
				Selector: ".alert-info",
				Extractor: Extractor{
					Types: []string{"text"},
				},
			},
		},
		Format: "json",
	}

	task = NewTask(badP)
	_, err = task.Parse()
	assert.Error(t, err, "errs.ErrNoPartOrSelectorProvided")

	//Bad output format
	badOF := Payload{
		Name: "BadOutputFormat",
		Request: fetch.Request{
			Type: "chrome",
			URL:  "http://testserver:12345",
		},
		Fields: []Field{
			{
				Name:     "Alert",
				Selector: ".alert-info",
				Extractor: Extractor{
					Types: []string{"text"},
				},
			},
		},
		//		PaginateResults: &paginateResults,
		Format: "BadOutputFormat",
	}
	task = NewTask(badOF)

	_, err = task.Parse()
	assert.Error(t, err, "invalid output format specified")

	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
}

func TestCSVEncode(t *testing.T) {
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	task := NewTask(CSVPayload)
	r, err := task.Parse()
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	resultFile := buf.Bytes()
	actual, err := ioutil.ReadFile(filepath.Join("./", string(resultFile)))
	assert.NoError(t, err)
	golden := filepath.Join("../testdata", "CSVEncode.csv")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
}

func TestXMLEncode(t *testing.T) {
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	task := NewTask(XMLPayload)
	r, err := task.Parse()
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	resultFile := buf.Bytes()
	actual, err := ioutil.ReadFile(filepath.Join("./", string(resultFile)))
	assert.NoError(t, err)
	golden := filepath.Join("../testdata", "XMLEncode.xml")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	//expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	//assert.Equal(t, expected, actual)

	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
}

//TestParseSwitchFetchers switch fetchers from type "base" to type "chrome" automatically in case of java scripts on a target web page need to be rendered.
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
	p := Payload{
		Name: "persons Table",
		Request: fetch.Request{
			Type: "base",
			URL:  "http://testserver:12345/persons/page-0",
		},
		Fields: []Field{
			{
				Name:     "Names",
				Selector: "#cards a",
				Extractor: Extractor{
					Types: []string{"text"},
				},
			},
		},
		Format: "json",
	}
	task := NewTask(p)
	r, err := task.Parse()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	os.RemoveAll("./diskv")
	os.RemoveAll("./results")
}

func TestScraper_partNames(t *testing.T) {
	s := Scraper{}
	s.Parts = []Part{
		{Name: "1"},
		{Name: "2"},
		{Name: "3"},
		{Name: "4"},
	}
	parts := s.partNames()
	assert.Equal(t, []string{"1", "2", "3", "4"}, parts)

}

func TestPayload_selectors(t *testing.T) {
	p1 := Payload{
		Fields: []Field{
			{Selector: "sel1"},
			{Selector: "sel2"},
			{Selector: "sel3"},
			{Selector: "sel4"},
		},
	}
	p2 := Payload{
		Fields: []Field{
			{},
			{},
			{},
			{},
		},
	}

	s1, err := p1.selectors()
	assert.NoError(t, err)
	assert.Equal(t, []string{"sel1", "sel2", "sel3", "sel4"}, s1)
	s2, err := p2.selectors()
	assert.Error(t, err)
	assert.Equal(t, []string(nil), s2)

}

func TestIntArrayToString(t *testing.T) {
	str := intArrayToString([]int{1, 2, 3, 4, 5}, ";")
	assert.Equal(t, "1;2;3;4;5", str)
}

func TestFloatArrayToString(t *testing.T) {
	str := floatArrayToString([]float64{1.1, 2.2, 3.3, 4.4, 5.5}, ";")
	assert.Equal(t, "1.1;2.2;3.3;4.4;5.5", str)
}
