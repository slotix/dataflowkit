package parser

type meta struct {
	Name string `json:"name" xml:"name"`
	URL  string `json:"url" xml:"url"`
}

type field struct {
	FieldName   string `json:"field_name"`
	CSSSelector string `json:"css_selector"`
}

//easyjson:json

type payload struct {
	meta
	
	Fields []field `json:"fields"`
}

//Parser structure stores input format and collections CSS Selectors
//easyjson:json
type Parser struct {
	Format string  `json:"format"`
	Collections []payload `json:"collections"`
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
	Element []collection `json:"collections"`
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
