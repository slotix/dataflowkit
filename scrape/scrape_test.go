package scrape

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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
	personsPayload, detailsPayload, JSONPayload, CSVPayload, XMLPayload, deepExtractPayload, pathDetailsPaginator Payload
	update                                                                                                        = flag.Bool("update", false, "update result files")
)

func init() {
	viper.Set("CHROME", "http://127.0.0.1:9222")
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	viper.Set("STORAGE_TYPE", "mongodb")
	viper.Set("RESULTS_DIR", "results")
	viper.Set("MAX_PAGES", 2)
	viper.Set("IGNORE_FETCH_DELAY", true)
	viper.Set("PAYLOAD_POOL_SIZE", 100)
	viper.Set("PAYLOAD_WORKERS_NUM", 50)
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
				Name:        "Names",
				CSSSelector: "#cards a",
				Attrs:       []string{"text", "href", "outerHtml"},
			},
			{
				Name:        "Images",
				CSSSelector: ".card-img-top",
				Attrs:       []string{"src", "alt", "width", "height"},
			},
		},
		Paginator: "nav:nth-child(4) :nth-child(2) .page-link",
		Format:    "json",
	}
	detailsPayload = Payload{
		Name: "persons details",
		Request: fetch.Request{
			Type: "chrome",
			URL:  "http://testserver:12345/persons/page-0",
		},
		Fields: []Field{
			{
				Name:        "Links",
				CSSSelector: "#cards a",
				Attrs:       []string{"text", "href"},
				Details: Payload{
					Fields: []Field{
						{
							Name:        "Number",
							CSSSelector: ".display-4",
							Attrs:       []string{"text"},
							Filters:     []Filter{Filter{"trim", ""}},
						},
						{
							Name:        "Name",
							CSSSelector: ".display-4",
							Attrs:       []string{"text"},
							Filters:     []Filter{Filter{"trim", ""}},
						},
						{
							Name:        "Company",
							CSSSelector: ".card-text:nth-child(3) .col-5",
							Attrs:       []string{"text"},
							Filters:     []Filter{Filter{"trim", ""}},
						},
						{
							Name:        "Phones",
							CSSSelector: ".col-10 span",
							Attrs:       []string{"text"},
							Filters:     []Filter{Filter{"trim", ""}},
						},
						{
							Name:        "Email",
							CSSSelector: ".card-text:nth-child(2) .col-5",
							Attrs:       []string{"text"},
							Filters:     []Filter{Filter{"trim", ""}},
						},
					},
				},
			},
			{
				Name:        "Count",
				CSSSelector: ".badge-primary",
				Attrs:       []string{"text"},
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
	JSONPayload = Payload{
		Name: "persons Cards",
		Request: fetch.Request{
			Type: "chrome",
			URL:  "http://testserver:12345/persons/page-0",
		},
		Fields: []Field{
			{
				Name:        "Names",
				CSSSelector: "#cards a",
				Attrs:       []string{"text", "href", "outerHtml"},
			},
			{
				Name:        "Images",
				CSSSelector: ".card-img-top",
				Attrs:       []string{"src", "alt", "width", "height"},
			},
		},
		Paginator: "nav:nth-child(4) :nth-child(2) .page-link",
		Format:    "",
	}
	CSVPayload = Payload{
		Name: "persons details",
		Request: fetch.Request{
			Type: "chrome",
			URL:  "http://testserver:12345/persons/page-0",
		},
		Fields: []Field{
			Field{
				Name:        "Meghan Reyes",
				CSSSelector: ".card:nth-child(3) a",
				Details: Payload{
					Fields: []Field{
						{
							Name:        "Name",
							CSSSelector: ".display-4",
							Attrs:       []string{"text"},
							Filters:     []Filter{Filter{"trim", ""}},
						},
						{
							Name:        "Phones",
							CSSSelector: ".col-10 span",
							Attrs:       []string{"text"},
							Filters:     []Filter{Filter{"trim", ""}},
						},
					},
				},
				Attrs: []string{"path"},
			},
		},
		IsPath: true,
		Format: "csv",
	}
	XMLPayload = Payload{
		Name: "persons details",
		Request: fetch.Request{
			Type: "chrome",
			URL:  "http://testserver:12345/persons/3",
		},
		Fields: []Field{
			{
				Name:        "Name",
				CSSSelector: ".display-4",
				Attrs:       []string{"text"},
				Filters:     []Filter{Filter{"trim", ""}},
			},
			{
				Name:        "Phones",
				CSSSelector: ".col-10 span",
				Attrs:       []string{"text"},
				Filters:     []Filter{Filter{"trim", ""}},
			},
		},
		Format: "xml",
	}
	deepExtractPayload = Payload{
		Name: "scrape.dataflowkit",
		Request: fetch.Request{
			URL:       "http://testserver:12345/",
			UserToken: "",
			Type:      "chrome",
		},
		Fields: []Field{
			{
				Name:        "Country_Button",
				CSSSelector: ".mr-5~ .mr-5+ .btn-primary",
				Details: Payload{
					IsPath: true,
					Fields: []Field{
						{
							Name:        "Countries",
							CSSSelector: ".list-group-item a",
							Details: Payload{
								IsPath: true,
								Fields: []Field{
									{
										Name:        "Cities",
										CSSSelector: ".list-group-item a",
										Details: Payload{
											IsPath:    true,
											Paginator: ".page-item:last-child .page-link",
											Fields: []Field{
												{
													Name:        "Persons",
													CSSSelector: "#cards a",
													Details: Payload{
														IsPath: false,
														Fields: []Field{
															{
																Name:        "Phone",
																CSSSelector: "span+ span",
																Attrs:       []string{"text"},
																Filters:     []Filter{Filter{"trim", ""}},
															},
															{
																Name:        "Country",
																CSSSelector: ".card-text:nth-child(1) a",
																Attrs:       []string{"text"},
																Filters:     []Filter{Filter{"trim", ""}},
															},
															{
																Name:        "City",
																CSSSelector: ".card-text+ .card-text a",
																Attrs:       []string{"text"},
																Filters:     []Filter{Filter{"trim", ""}},
															},
															{
																Name:        "Title",
																CSSSelector: ".display-4",
																Attrs:       []string{"text"},
																Filters:     []Filter{Filter{"trim", ""}},
															},
														},
													},
													Attrs:   []string{"path"},
													Filters: []Filter{Filter{"trim", ""}},
												},
											},
										},
										Attrs:   []string{"path"},
										Filters: []Filter{Filter{"trim", ""}},
									},
								},
							},
							Attrs:   []string{"path"},
							Filters: []Filter{Filter{"trim", ""}},
						},
					},
				},
				Attrs:   []string{"path"},
				Filters: []Filter{Filter{"trim", ""}},
			},
		},
		Format: "json",
		IsPath: true,
	}
	pathDetailsPaginator = Payload{
		Name: "details paginator",
		Request: fetch.Request{
			URL:       "http://testserver:12345/country/United%20States",
			UserToken: "",
			Type:      "",
			Actions:   "",
		},
		Fields: []Field{
			Field{
				Name:        "selector1",
				CSSSelector: ".list-group-item a",
				Details: Payload{
					Name: "selector1details",
					Request: fetch.Request{
						URL:       "http://testserver:12345/country/United%20States/city/San%20Jose",
						UserToken: "",
						Type:      "",
						Actions:   "",
					},
					Fields: []Field{
						Field{
							Name:        "selector1",
							CSSSelector: ".badge-primary",
							Attrs:       []string{"text"},
							Filters: []Filter{
								Filter{Name: "trim"},
							},
						},
						Field{
							Name:        "selector2",
							CSSSelector: "#cards a",
							Attrs:       []string{"href", "text"},
							Filters: []Filter{
								Filter{Name: "trim"},
							},
						},
					},
					Paginator: ".active~ .page-item+ .page-item .page-link",
					Format:    "",
					IsPath:    false,
				},
				Attrs: []string{"path"},
				Filters: []Filter{
					Filter{Name: "trim"},
				},
			},
		},
		Paginator: "",
		Format:    "json",
		IsPath:    true,
	}
}

func TestNewTask(t *testing.T) {
	task := NewTask()
	assert.NotEmpty(t, task.storage)
}

func TestParseDetails(t *testing.T) {
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	//JSON details output
	ctx := context.Background()
	detailsPayload.Format = "json"
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, detailsPayload)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	str := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)

	resultFile := str["Output file"]
	actual, err := ioutil.ReadFile(filepath.Join("./", resultFile.(string)))
	assert.NoError(t, err)
	golden := filepath.Join("../testdata/scrape", "details.json")
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

	task.storage.DeleteAll()
	os.RemoveAll("./results")
}

func TestParse(t *testing.T) {
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	//JSON output
	ctx := context.Background()
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, personsPayload)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	str := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)
	resultFile := str["Output file"]
	actual, err := ioutil.ReadFile(filepath.Join("./", resultFile.(string)))
	assert.NoError(t, err)
	//t.Log(string(got))
	golden := filepath.Join("../testdata/scrape", "result.json")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	//xml
	personsPayload.Format = "xml"
	task = NewTask()
	r, err = task.Parse(ctx, personsPayload)
	assert.NoError(t, err)
	buf = new(bytes.Buffer)
	buf.ReadFrom(r)
	str = make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)
	resultFile = str["Output file"]
	actual, err = ioutil.ReadFile(filepath.Join("./", resultFile.(string)))
	assert.NoError(t, err)
	golden = filepath.Join("../testdata/scrape", "result.xml")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	_, err = ioutil.ReadFile(golden)
	assert.NoError(t, err)
	//todo: order of fields in both xml files is not identical. So it is not possible to compare them easily.
	//assert.Equal(t, expected, actual)

	task.storage.DeleteAll()
	os.RemoveAll("./results")
}

func TestParseErrs(t *testing.T) {
	///// No selectors
	badP := Payload{
		Name: "No Selectors",
		Request: fetch.Request{
			URL: "http://127.0.0.1:12345",
		},
		Format: "json",
	}

	ctx := context.Background()
	task := NewTask()
	_, err := task.Parse(ctx, badP)
	assert.Error(t, err, "Bad payload: No fields to scrape")

	///// ErrNoPartOrSelectorProvided
	badP = Payload{
		Name: "ErrNoPartOrSelectorProvided",
		Request: fetch.Request{
			URL: "http://127.0.0.1:12345",
		},
		Fields: []Field{
			{
				Name:        "Alert",
				CSSSelector: "",
				Attrs:       []string{"text"},
			},
			{
				Name:        "",
				CSSSelector: ".alert-info",
				Attrs:       []string{"text"},
			},
		},
		Format: "json",
	}

	task = NewTask()
	_, err = task.Parse(ctx, badP)
	assert.Error(t, err, "Bad payload: Field 0 has no css selector")

	badP.Fields[0].CSSSelector = "selector"
	task = NewTask()
	_, err = task.Parse(ctx, badP)
	assert.Error(t, err, "Bad payload: Field 1 has no name")

	//Bad output format
	badOF := Payload{
		Name: "BadOutputFormat",
		Request: fetch.Request{
			Type: "chrome",
			URL:  "http://testserver:12345",
		},
		Fields: []Field{
			{
				Name:        "Alert",
				CSSSelector: ".alert-info",
				Attrs:       []string{"text"},
			},
		},
		//		PaginateResults: &paginateResults,
		Format: "BadOutputFormat",
	}
	task = NewTask()

	_, err = task.Parse(ctx, badOF)
	assert.Error(t, err, "Bad payload: Unsupported output format BadOutputFormat")
}

func TestJSONEncode(t *testing.T) {
	JSONPayload.Format = "json"
	os.RemoveAll("./results")

	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	ctx := context.Background()
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, JSONPayload)

	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	str := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)
	resultFile := str["Output file"]
	filename := resultFile.(string)
	assert.Equal(t, filename[len(filename)-4:], "json")
	actual, err := ioutil.ReadFile(filepath.Join("./", resultFile.(string)))
	assert.NoError(t, err)
	//t.Log(string(got))
	golden := filepath.Join("../testdata/scrape", "encode.json")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	task.storage.DeleteAll()
	os.RemoveAll("./results")
}

func TestJSONLEncode(t *testing.T) {
	JSONPayload.Format = "jsonl"
	os.RemoveAll("./results")

	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	ctx := context.Background()
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, JSONPayload)

	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	str := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)
	resultFile := str["Output file"]
	filename := resultFile.(string)
	assert.Equal(t, filename[len(filename)-5:], "jsonl")
	actual, err := ioutil.ReadFile(filepath.Join("./", resultFile.(string)))
	assert.NoError(t, err)
	golden := filepath.Join("../testdata/scrape", "encode.jsonl")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	task.storage.DeleteAll()
	os.RemoveAll("./results")
}

func TestCSVEncode(t *testing.T) {
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	ctx := context.Background()
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, CSVPayload)

	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	str := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)
	resultFile := str["Output file"]
	filename := resultFile.(string)
	assert.Equal(t, filename[len(filename)-3:], "csv")
	actual, err := ioutil.ReadFile(filepath.Join("./", resultFile.(string)))
	assert.NoError(t, err)
	golden := filepath.Join("../testdata/scrape", "encode.csv")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	task.storage.DeleteAll()
	os.RemoveAll("./results")
}

// TODO: TestXMLEncode && TestXLSEncode
//  XML - order of fields in both xml files is not identical. Structure compare required.
//  XLS - structure compare required because of metadata

func TestXMLEncode(t *testing.T) {
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	ctx := context.Background()
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, XMLPayload)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	str := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)
	resultFile := str["Output file"]
	actual, err := ioutil.ReadFile(filepath.Join("./", resultFile.(string)))
	assert.NoError(t, err)
	golden := filepath.Join("../testdata/scrape", "encode.xml")
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
	}
	//expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	//assert.Equal(t, expected, actual)

	task.storage.DeleteAll()
	os.RemoveAll("./results")
}

//TestParseSwitchFetchers switch fetchers from type "base" to type "chrome" automatically in case of java scripts on a target web page need to be rendered.
func TestParseSwitchFetchers(t *testing.T) {
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
				Name:        "Names",
				CSSSelector: "#cards a",
				Attrs:       []string{"text"},
			},
		},
		Format: "json",
	}
	ctx := context.Background()
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, p)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	task.storage.DeleteAll()
	os.RemoveAll("./results")
}

func TestIntArrayToString(t *testing.T) {
	str := intArrayToString([]int{1, 2, 3, 4, 5}, ";")
	assert.Equal(t, "1;2;3;4;5", str)
}

func TestFloatArrayToString(t *testing.T) {
	str := floatArrayToString([]float64{1.1, 2.2, 3.3, 4.4, 5.5}, ";")
	assert.Equal(t, "1.1;2.2;3.3;4.4;5.5", str)
}

func TestGZipJSONEncode(t *testing.T) {
	JSONPayload.Format = "json"
	JSONPayload.Compressor = GZIP_COMPRESS
	os.RemoveAll("./results")

	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	ctx := context.Background()
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, JSONPayload)

	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	str := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)
	resultFile := str["Output file"]
	filename := resultFile.(string)
	assert.Equal(t, filename[len(filename)-2:], GZIP_COMPRESS)

	actual := filepath.Join("./", resultFile.(string))

	fo, err := os.OpenFile(actual, os.O_RDONLY, 0660)
	assert.NoError(t, err)
	gr, err := gzip.NewReader(fo)
	assert.NoError(t, err)
	bb, err := ioutil.ReadAll(gr)
	assert.NoError(t, err)
	assert.NoError(t, fo.Close())

	//t.Log(string(got))
	golden := filepath.Join("../testdata/scrape", "encode.json")
	expected, err := ioutil.ReadFile(golden)
	assert.NoError(t, err)
	assert.Equal(t, expected, bb)

	JSONPayload.Compressor = ""
	task.storage.DeleteAll()
	os.RemoveAll("./results")
}

func TestFilters(t *testing.T) {
	filter := Filter{"regex", ""}
	res, err := filter.Apply("")
	assert.Error(t, err, "Data source is empty")

	res, err = filter.Apply("abc")
	assert.Error(t, err, "No regex given")

	filter.Param = "(\\d)(\\d+)"
	res, err = filter.Apply("1234")
	assert.Error(t, err, "Regex filter doesn't support subexpressions")

	filter.Param = "\\d"
	res, err = filter.Apply("1234")
	assert.NoError(t, err)
	assert.Equal(t, "1;2;3;4;", res)
}

func TestPathPaginator(t *testing.T) {
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	ctx := context.Background()
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, pathDetailsPaginator)
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	str := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)
	resultFile := str["Output file"].(string)

	actualText, err := ioutil.ReadFile(filepath.Join("./", resultFile))
	assert.NoError(t, err)

	var actualJSON []map[string]string
	err = json.Unmarshal([]byte(actualText), &actualJSON)
	assert.NoError(t, err)

	assert.Equal(t, 25, len(actualJSON))

	task.storage.DeleteAll()
	os.RemoveAll("./results")
}

func TestPathParse(t *testing.T) {
	viper.Set("MAX_PAGES", 10)
	os.RemoveAll("./results")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	ctx := context.Background()
	task := NewTask()
	task.storage.DeleteAll()
	r, err := task.Parse(ctx, deepExtractPayload)
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	str := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &str)
	assert.NoError(t, err)
	resultFile := str["Output file"].(string)

	actualText, err := ioutil.ReadFile(filepath.Join("./", resultFile))
	assert.NoError(t, err)

	var actualJSON []map[string]string
	err = json.Unmarshal([]byte(actualText), &actualJSON)
	assert.NoError(t, err)

	// 100 - expected persons after Parse of deepExtractPayload
	assert.Equal(t, 100, len(actualJSON))

	for _, item := range actualJSON {
		value, exists := item["Phone_text"]
		assert.Equal(t, true, exists)
		assert.NotEqual(t, "", value)

		value, exists = item["Country_text"]
		assert.Equal(t, true, exists)
		assert.NotEqual(t, "", value)

		value, exists = item["City_text"]
		assert.Equal(t, true, exists)
		assert.NotEqual(t, "", value)

		value, exists = item["Title_text"]
		assert.Equal(t, true, exists)
		assert.NotEqual(t, "", value)
	}
	task.storage.DeleteAll()
	os.RemoveAll("./results")
}
