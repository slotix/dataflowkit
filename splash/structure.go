package splash

import (
	"net/http"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
)

type Request struct {
	//URL is required to be passed to Fetch Endpoint
	URL string `json:"url"`
	//Params used for passing formdata to LUA script
	Params string `json:"params,omitempty"`
	//Cookies contain cookies to be added to request
	Cookies string `json:"cookie,omitempty"`
	Func    string `json:"func,omitempty"`
	//SplashWait - time in seconds to wait until js scripts loaded. Sometimes wait parameter should be set to more than default 0,5. It allows to finish js scripts execution on a web page.
	//	SplashWait float64 `json:"wait,omitempty"`
}

//Cookie - Custom Cookie struct is used to avoid problems with unmarshalling data with invalid Expires field which has time.Time type for original http.Cookie struct.
//For some domains like http://yahoo.com it is easier to unmarshal Expires as string
type Cookie struct {
	http.Cookie
	Expires string
}

//SResponse returned by Splash as a part of Response
//It is needed to be passed to caching middleware
type SResponse struct {
	Headers     interface{} `json:"headers"`
	HeadersSize int         `json:"headersSize"`
	Cookies     []Cookie    `json:"cookies"`
	Ok          bool        `json:"ok"`
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

//SRequest returned by Splash as a part of Response
//It is needed to be passed to caching middleware
type SRequest struct {
	Method      string      `json:"method"`
	Headers     interface{} `json:"headers"`
	HeadersSize int         `json:"headersSize"`
	Cookies     []Cookie    `json:"cookies"`
	URL         string      `json:"url"`
	HTTPVersion string      `json:"httpVersion"`
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
	//If Error is not nil all other fields are nil
	Error             string     `json:"error,omitempty"`
	Request           *SRequest  `json:"request"`
	Response          *SResponse `json:"response"`
	ReasonsNotToCache []cacheobject.Reason
	//Cacheable bool
	Expires time.Time //how long object stay in a cache before Splash fetcher forwards another request to an origin.
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
