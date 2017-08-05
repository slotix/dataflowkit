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
	URL string `json:"url"`
	//Params used for passing formdata to LUA script
	Params string `json:"params,omitempty"`
	//Cookies contain cookies to be added to request
	Cookies string `json:"cookie,omitempty"`
	Func    string `json:"func,omitempty"`
	//SplashWait - time in seconds to wait until js scripts loaded. Sometimes wait parameter should be set to more than default 0,5. It allows to finish js scripts execution on a web page.
	SplashWait float64 `json:"wait,omitempty"`
}

//Cookie - Custom Cookie struct is used to avoid problems whith unmarshalling data with invalid Expires field which has time.Time type for original http.Cookie struct.
//For some domains like http://yahoo.com it is easier to unmarshal Expires as string
type Cookie struct {
	http.Cookie
	Expires string // optional
}

//SResponse returned by Splash as a part of Response
//It is needed to be passed to caching middleware
type SResponse struct {
	Headers interface{} `json:"headers"`
	//Headers     []Header      `json:"headers"`
	HeadersSize int      `json:"headersSize"`
	Cookies     []Cookie `json:"cookies"`
	//Cookies []Cookie `json:"cookies"`
	Ok      bool `json:"ok"`
	Content struct {
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

//SRequest returned by Splash as a part of Response
//It is needed to be passed to caching middleware
type SRequest struct {
	Method  string      `json:"method"`
	Headers interface{} `json:"headers"`
	//Headers     []Header      `json:"headers"`
	HeadersSize int      `json:"headersSize"`
	Cookies     []Cookie `json:"cookies"`
	//Cookies     []Cookie `json:"cookies"`
	URL         string `json:"url"`
	HTTPVersion string `json:"httpVersion"`
	QueryString []struct {
		Value string `json:"value"`
		Name  string `json:"name"`
	} `json:"queryString"`
	BodySize int `json:"bodySize"`
}

//Response returned by Splash
//It includes html body, response, request
type Response struct {
	URL  string `json:"url"`
	HTML string `json:"html"`
	//Error is returned in case of an error, f.e. "http404".
	//All other fields will be nil if Error is not nil
	Error string `json:"error,omitempty"`
	//Cookies             []Cookie   `json:"cookies"`
	Request             *SRequest  `json:"request"`
	Response            *SResponse `json:"response"`
	Cacheable           bool
	CacheExpirationTime int64
}
//PingResponse returned by Splash _ping  endpoint
type PingResponse struct {
	Maxrss int    `json:"maxrss"`
	Status string `json:"status"`
}

type gcResponse struct {
	CachedArgsRemoved  int    `json:"cached_args_removed"`
	PyObjectsCollected int    `json:"pyobjects_collected"`
	Status             string `json:"status"`
}
