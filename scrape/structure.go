package scrape

import (
	"time"
)

type Extractor struct {
	Type   string      `json:"type"`
	Params interface{} `json:"params"`
}

type field struct {
	Name     string `json:"name" validate:"required"`
	Selector string `json:"selector" validate:"required"`
	//Count     int       `json:"count"`
	Extractor Extractor `json:"extractor"`
	Details   *Payload  `json:"details"`
}

type paginator struct {
	Selector  string `json:"selector"`
	Attribute string `json:"attr"`
	MaxPages  int    `json:"maxPages"`
}

type Payload struct {
	Name string `json:"name" validate:"required"`
	//Request             splash.Request `json:"request"`
	Request             interface{}   `json:"request"`
	Fields              []field       `json:"fields"`
	//PayloadMD5 encodes payload content to MD5. It is used for generating file name to be stored.
	PayloadMD5          []byte        `json:"payloadMD5"`
	Format              string        `json:"format"`
	Paginator           paginator     `json:"paginator"`
	PaginateResults     *bool         `json:"paginateResults"`
	FetchDelay          time.Duration `json:"fetchDelay"`
	RandomizeFetchDelay *bool         `json:"randomizeFetchDelay"`
	RetryTimes          int           `json:"retryTimes"`
}
