package parser

type meta struct {
	Name string `json:"name" xml:"name" validate:"required"`
	URL  string `json:"url" xml:"url" validate:"required"`
}

type field struct {
	Name        string `json:"name" validate:"required"`
	CSSSelector string `json:"css" validate:"required"`
	Type        int    `json:"type"`
	Count       int    `json:"count"`
	Details     payload `json:"-" validate:"-"`
	//Regex       string `json:"regex"`
	//FieldType   string `json:"type"`
}

//easyjson:json

type payload struct {
	meta
	Fields []field `json:"fields" validate:"gt=0"` //number of fields >0 
}

//Parser structure stores input format and collections CSS Selectors
//easyjson:json
type Parser struct {
	Format     string    `json:"format"`
	Payloads   []payload `json:"collections"`
	PayloadMD5 []byte    `json:"payloadMD5"`
}

//easyjson:json
type collection struct {
	meta
	Items     []interface{} `json:"items"`
	Fields    []string      `json:"-"`
	Count     int           `json:"count"`
	CreatedAt int64         `json:"time"`
}

//Collections structure stores output data
//easyjson:json
type Collections struct {
	//	Format string `json:"format"`
	Collections []*collection `json:"collections"`
}

//CSVTable structure stores output data
//easyjson:json
type CSVTable struct {
	Content string `json:"table"`
	URL     string `json:"url"`
}

//easyjson:json
type CSVTableCollection struct {
	Tables []CSVTable `json:"tables"`
}
