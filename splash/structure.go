package splash

import (
	"net/http"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
)

//Request struct is entry point which is initially filled from Payload
//it is passed to Splash server to fetch web content.
type Request struct {
	//URL holds the URL address of the web page to be downloaded.
	URL string `json:"url"`
	// Params is a string value for passing formdata parameters.
	//
	// For example it may be used for processing pages which require authentication
	//
	// Example:
	//
	// "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=userpassword&rememberMe=1"
	//
	Params string `json:"params,omitempty"`
	// Cookies contain cookies to be added to request  before sending it to browser.
	//
	// It may be used for processing pages after initial authentication. In the first step formdata with auth info is passed to a web page.
	//
	// Response object headers may contain an Object like
	//
	// name: "Set-Cookie"
	//
	// value: "session_id=29d7b97879209ca89316181ed14eb01f; path=/; httponly"
	//
	// These cookie should be passed to the next pages on the same domain.
	//
	// "session_id", "29d7b97879209ca89316181ed14eb01f", "/", domain="example.com"
	//
	//Cookies string `json:"cookie,omitempty"`
	Cookies []*http.Cookie `json:"cookie,omitempty"`
	//LUA Splash custom script
	LUA string `json:"lua,omitempty"`
}

//Cookie - Custom Cookie struct is used to avoid problems with unmarshalling data with invalid Expires field which has time.Time type for original http.Cookie struct.
//For some domains like http://yahoo.com it is easier to unmarshal Expires as string
type Cookie struct {
	http.Cookie
	//Expires string
}

//SResponse returned by Splash as a part of Response
//It is needed to be passed to caching middleware
type SResponse struct {
	Headers     interface{} `json:"headers"`
	HeadersSize int         `json:"headersSize"`
	//Cookies     []Cookie    `json:"cookies"`
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
	URL         string      `json:"url"`
	Method      string      `json:"method"`
	Headers     interface{} `json:"headers"`
	HeadersSize int         `json:"headersSize"`
	Cookies     []Cookie    `json:"cookies"`
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
	//URL  string `json:"url"`
	HTML string `json:"html"`
	//Error is returned in case of an error, f.e. "http404".
	//If Error is not nil all other fields are nil
	Error    string     `json:"error,omitempty"`
	Request  *SRequest  `json:"request"`
	Response *SResponse `json:"response"`
	Cookies  []Cookie   `json:"cookies"`
	//ReasonsNotToCache is an array of reasons why a response should not be cached.
	ReasonsNotToCache []cacheobject.Reason
	//Expires - How long object stay in a cache before Splash fetcher forwards another request to an origin.
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
