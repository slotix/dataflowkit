package cache

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/slotix/dataflowkit/downloader"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "cache: ", log.Lshortfile)
}

func generateHTTPHeaders(dHeaders downloader.Headers) (header http.Header) {
	header = make(map[string][]string)
	for _, h := range dHeaders {
		var str []string
		str = append(str, h.Value)
		header[h.Name] = str
	}	
	return header
}

//Cacheable check if resource is cacheable
func Cacheable(resp *downloader.SplashResponse) (rv cacheobject.ObjectResults) {
	
	respHeader := generateHTTPHeaders(resp.Response.Headers)
	reqHeader := generateHTTPHeaders(resp.Request.Headers)

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
		RespDirectives:         resDir,
		RespHeaders:            respHeader,
		RespStatusCode:         resp.Response.Status,
		RespExpiresHeader:      expiresHeader,
		RespDateHeader:         dateHeader,
		RespLastModifiedHeader: lastModifiedHeader,

		ReqDirectives: reqDir,
		ReqHeaders:    reqHeader,
		ReqMethod:     resp.Request.Method,

		NowUTC: time.Now().UTC(),
	}

	rv = cacheobject.ObjectResults{}
	cacheobject.CachableObject(&obj, &rv)
	cacheobject.ExpirationObject(&obj, &rv)
	return rv
}
