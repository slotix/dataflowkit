package fetch

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/slotix/dataflowkit/errs"
)

//BaseFetcherRequest struct collects requests information used by BaseFetcher
type BaseFetcherRequest struct {
	URL    string //URL to be retrieved 
	Method string //HTTP method : GET, POST
}

//BaseFetcherResponse struct groups Response data together after retrieving it by BaseFetcher
type BaseFetcherResponse struct {
	//Response is used for determining Cacheable and Expires values. It should be omited when marshaling to intermediary cache.
	Response   *http.Response `json:"-"`
	HTML       []byte         `json:"html"`
	//Cacheable checks if html page is cacheable. If no then it will be downloaded every time it is requested.  
	Cacheable  bool
	//Expires - How long object stay in a cache before Splash fetcher forwards another request to an origin.
	Expires    time.Time 
	StatusCode int
	Status     string
}

//MarshalJSON customizes marshaling of http.Response.Body which has type io.ReadCloser. It cannot be marshaled with standard Marshal method without casting to []byte.
//http://choly.ca/post/go-json-marshalling/
func (r *BaseFetcherResponse) MarshalJSON() ([]byte, error) {
	type Alias BaseFetcherResponse
	body, err := ioutil.ReadAll(r.Response.Body)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&struct {
		HTML []byte `json:"-"`
		*Alias
	}{
		HTML:  body,
		Alias: (*Alias)(r),
	})
}

//setCacheInfo check if resource is cacheable
//r.Cacheable and r.CacheExpirationTime fields are filled inside this func
func (r *BaseFetcherResponse) SetCacheInfo() {
	respHeader := r.Response.Header
	reqHeader := r.Response.Request.Header

	reqDir, err := cacheobject.ParseRequestCacheControl(reqHeader.Get("Cache-Control"))
	if err != nil {
		logger.Printf(err.Error())
	}
	resDir, err := cacheobject.ParseResponseCacheControl(respHeader.Get("Cache-Control"))
	if err != nil {
		logger.Printf(err.Error())
	}
	//logger.Println(respHeader)
	expiresHeader, _ := http.ParseTime(respHeader.Get("Expires"))
	dateHeader, _ := http.ParseTime(respHeader.Get("Date"))
	lastModifiedHeader, _ := http.ParseTime(respHeader.Get("Last-Modified"))
	obj := cacheobject.Object{
		//	CacheIsPrivate:         false,
		RespDirectives: resDir,
		RespHeaders:    respHeader,
		RespStatusCode: r.Response.StatusCode,

		RespExpiresHeader:      expiresHeader,
		RespDateHeader:         dateHeader,
		RespLastModifiedHeader: lastModifiedHeader,

		ReqDirectives: reqDir,
		ReqHeaders:    reqHeader,
		ReqMethod:     r.Response.Request.Method,
		NowUTC:        time.Now().UTC(),
	}

	rv := cacheobject.ObjectResults{}
	cacheobject.CachableObject(&obj, &rv)
	cacheobject.ExpirationObject(&obj, &rv)

	//Check if it is cacheable

	if len(rv.OutReasons) == 0 {
		r.Cacheable = true
		if rv.OutExpirationTime.IsZero() {
			//if time is zero than set it to current time plus 24 hours.
			r.Expires = time.Now().UTC().Add(time.Hour * 24)
		} else {
			r.Expires = rv.OutExpirationTime
		}
		logger.Println("Current Time: ", time.Now().UTC())
		logger.Println(r.Response.Request.URL, r.GetExpires())
	} else {
		//if resource is not cacheable set expiration time to the current time.
		//This way web page will be downloaded every time.
		r.Expires = time.Now().UTC()
	}

	debug := false
	if debug {
		if rv.OutErr != nil {
			logger.Println("Errors: ", rv.OutErr)
		}
		if rv.OutReasons != nil {
			logger.Println("Reasons to not cache: ", rv.OutReasons)
		}
		if rv.OutWarnings != nil {
			logger.Println("Warning headers to add: ", rv.OutWarnings)
		}
		logger.Println("Expiration: ", rv.OutExpirationTime.String())
	}
	//return rv
}

//GetExpires returns Expires field of Response
func (r BaseFetcherResponse) GetExpires() time.Time {
	return r.Expires
}

//GetExpires returns Cacheable field of Response
func (r BaseFetcherResponse) GetCacheable() bool {
	return r.Cacheable
}

//GetExpires returns URL to be downloaded  
func (r BaseFetcherRequest) GetURL() string {
	return strings.TrimSpace(strings.TrimRight(r.URL, "/"))
}

//Validate validates request to be send, prior to sending.
func (r BaseFetcherRequest) Validate() error {
	reqURL := strings.TrimSpace(r.URL)
	if _, err := url.ParseRequestURI(reqURL); err != nil {
		return &errs.BadRequest{err}
	}
	return nil
}
