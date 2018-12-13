package scrape

import (
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/slotix/dataflowkit/fetch"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetCommonAncestor(t *testing.T) {
	assert := assert.New(t)
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()
	fetcher, _ := fetch.NewHTTPClient("127.0.0.1:8000")
	selectors := []string{}
	req := fetch.Request{
		URL:  "http://testserver:12345/persons/page-0",
		Type: "chrome",
	}
	time.Sleep(500 * time.Millisecond)
	content, _ := fetcher.Fetch(req)
	doc, _ := goquery.NewDocumentFromReader(content)
	_, err := getCommonAncestor(doc.Selection, selectors)
	assert.Error(err, "it should return error")
	selectors = []string{".card-img-top", "#cards a"}
	sel, err := getCommonAncestor(doc.Selection, selectors)
	assert.NoError(err, "Error should be nil")
	assert.NotNil(sel, "Selectors not nil")
	suc := func() bool { return sel.Length() > 0 }
	assert.Condition(suc)
	selectors = []string{".not-exist-element", "#cards a"}
	sel, err = getCommonAncestor(doc.Selection, selectors)
	assert.NoError(err, "Error should be nil")
	assert.NotNil(sel, "Selectors not nil")
	assert.Condition(suc)

	selectors = []string{".not-exist-element", ".not-exist-element"}
	_, err = getCommonAncestor(doc.Selection, selectors)
	assert.Error(err, "Should return error")
}
