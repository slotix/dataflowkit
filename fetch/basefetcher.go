package fetch

import (
	"net/http"
	"strings"
	"time"

	"github.com/pquerna/cachecontrol"
	"github.com/pquerna/cachecontrol/cacheobject"
)

//BaseFetcherRequest struct collects requests information used by BaseFetcher
type BaseFetcherRequest struct {
	//URL to be retrieved
	URL string
	//HTTP method : GET, POST
	Method string
}

//BaseFetcherResponse struct groups Response data together after retrieving it by BaseFetcher
type BaseFetcherResponse struct {
	//Response is used for determining Cacheable and Expires values. It should be omitted when marshaling to intermediary cache.
	Response *http.Response `json:"-"`
	//URL represents the final URL after all redirects. Response.Request.URL.String()
	URL string
	//HTML Content of fetched page
	HTML []byte `json:"html"`
	//ReasonsNotToCache is an array of reasons why a response should not be cached.
	ReasonsNotToCache []cacheobject.Reason
	//Expires - How long object stay in a cache before Splash fetcher forwards another request to an origin.
	Expires time.Time
	//Status code returned by fetcher
	StatusCode int
	//Status returned by fetcher
	Status string
}

//MarshalJSON customizes marshaling of http.Response.Body which has type io.ReadCloser. It cannot be marshaled with standard Marshal method without casting to []byte.
//http://choly.ca/post/go-json-marshalling/
/* func (r *BaseFetcherResponse) MarshalJSON() ([]byte, error) {
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
} */

// SetCacheInfo checks if resource is cacheable.
// Respource is cachable if length of ReasonsNotToCache is zero.
// ReasonsNotToCache and Expires values are filled here
func (r *BaseFetcherResponse) SetCacheInfo() {
	reasons, expires, err := cachecontrol.CachableResponse(r.Response.Request, r.Response, cachecontrol.Options{})
	if err != nil {
		logger.Errorln(err)
	}
	if expires.IsZero() {
		//if time is zero than set it to current time plus 24 hours.
		r.Expires = time.Now().UTC().Add(time.Hour * 24)
	} else {
		r.Expires = expires
	}
	r.ReasonsNotToCache = reasons
}

//GetURL returns URL after all redirects
func (r BaseFetcherResponse) GetURL() string {
	return r.URL
}

//GetHTML returns HTML content
func (r BaseFetcherResponse) GetHTML() []byte {
	return r.HTML
}

//GetExpires returns Response Expires value.
func (r BaseFetcherResponse) GetExpires() time.Time {
	return r.Expires
}

//GetReasonsNotToCache returns an array of reasons why a response should not be cached.
func (r BaseFetcherResponse) GetReasonsNotToCache() []cacheobject.Reason {
	return r.ReasonsNotToCache
}

//GetURL returns URL to be fetched
func (req BaseFetcherRequest) GetURL() string {
	return strings.TrimRight(strings.TrimSpace(req.URL), "/")
}


