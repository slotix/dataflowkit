package scrape

import "github.com/slotix/dataflowkit/splash"

type Extractor struct {
	Type   string      `json:"type"`
	Params interface{} `json:"params"`
}

type field struct {
	Name      string    `json:"name" validate:"required"`
	Selector  string    `json:"selector" validate:"required"`
	Count     int       `json:"count"`
	Details   Payload   `json:"-" validate:"-"`
	Extractor Extractor `json:"extractor"`
}

type paginator struct {
	Selector  string `json:"selector"`
	Attribute string `json:"attr"`
	MaxPages  int    `json:"maxPages"`
}

type Payload struct {
	Name             string         `json:"name" validate:"required"`
	Request          splash.Request `json:"request"`
	Fields           []field        `json:"fields" validate:"gt=0"`
	Paginator        paginator      `json:"paginator"`
	PayloadMD5       []byte         `json:"payloadMD5"`
	Format           string         `json:"format"`
	PaginatedResults bool           `json:"paginatedResults"`
}
