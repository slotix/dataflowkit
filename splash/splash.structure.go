package splash

import (
	"net/http"
)

type Connection struct {
	Host            string
	User            string
	Password        string
	Timeout         int
	ResourceTimeout int
	LUAScript       string
}

type Request struct {
	URL     string  `json:"url"`
	Params  string  `json:"params,omitempty"` //params used for passing formdata to LUA script
	Cookies string  `json:"cookie,omitempty"` //add cookie to request
	Func    string  `json:"func,omitempty"`
	Wait    float64 `json:"wait,omitempty"` //Time in seconds to wait until js scripts loaded. Sometimes wait parameter should be set to more than default 0,5. It allows to finish js scripts execution on a web page.
}

//SResponse returned by splash as a part of Response
//It is needed to be passed to caching middleware
type SResponse struct {
	Headers     interface{}   `json:"headers"`
	Cookies     []http.Cookie `json:"cookies"`
	HeadersSize int           `json:"headersSize"`
	Ok          bool          `json:"ok"`
	Content     struct {
		Text     string `json:"text"`
		MimeType string `json:"mimeType"`
		Size     int    `json:"size"`
		Encoding string `json:"encoding"`
	} `json:"content"`
	Status      int    `json:"status"`
	URL         string `json:"url"`
	HTTPVersion string `json:"httpVersion"`
	StatusText  string `json:"statusText"`
	RedirectURL string `json:"redirectURL"`
}

//SRequest returned by splash as a part of Response
//It is needed to be passed to caching middleware
type SRequest struct {
	Cookies     []http.Cookie `json:"cookies"`
	Method      string        `json:"method"`
	HeadersSize int           `json:"headersSize"`
	URL         string        `json:"url"`
	HTTPVersion string        `json:"httpVersion"`
	QueryString []struct {
		Value string `json:"value"`
		Name  string `json:"name"`
	} `json:"queryString"`
	Headers interface{} `json:"headers"`

	BodySize int `json:"bodySize"`
}

//Response returned by splash
//It includes html body, response, request
type Response struct {
	HTML                string        `json:"html"`
	Reason              string        `json:"reason"`
	Cookies             []http.Cookie `json:"cookies"`
	Request             SRequest      `json:"request"`
	Response            SResponse     `json:"response"`
	Cacheable           bool
	CacheExpirationTime int64
}

type PingResponse struct {
	Maxrss int    `json:"maxrss"`
	Status string `json:"status"`
}
