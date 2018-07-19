package fetch

import (
	"net/http"
	"time"

	"github.com/pquerna/cachecontrol"
	"github.com/pquerna/cachecontrol/cacheobject"
)

//BaseFetcherResponse struct groups Response data together after retrieving it by BaseFetcher
type BaseFetcherResponse struct {
	//Response is used for determining Cacheable and Expires values. It should be omitted when marshaling to intermediary cache.
	Response *http.Response `json:"-"`
	//URL represents the final URL after all redirects. Response.Request.URL.String()
	URL string
	//HTML Content of fetched page
	HTML string `json:"html"`
	//ReasonsNotToCache is an array of reasons why a response should not be cached.
	ReasonsNotToCache []cacheobject.Reason
	//Expires - How long object stay in a cache before Splash fetcher forwards another request to an origin.
	Expires time.Time
	//Status code returned by fetcher
	StatusCode int
	//Status returned by fetcher
	Status string
}

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