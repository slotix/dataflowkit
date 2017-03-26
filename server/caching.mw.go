package server

import (
	"encoding/json"
	"errors"

	"fmt"

	"github.com/slotix/dataflowkit/cache"
	"github.com/slotix/dataflowkit/downloader"
	"github.com/slotix/dataflowkit/parser"
)

var errRedisSet = errors.New("Redis. Set Value failed")

func cachingMiddleware() ServiceMiddleware {
	return func(next ParseService) ParseService {
		return cachemw{next}
	}
}

type cachemw struct {
	ParseService
}

var redisCon cache.RedisConn 
func init(){
	redisURL := "localhost:6379"
	redisPassword := ""
	redisCon = cache.NewRedisConn(redisURL, redisPassword, "", 0)
}

//Cache-Control -> private, no-store -> disallowed storing in cache
//https://redbot.org/
//https://www.mnot.net/cache_docs/

//Expires HTTP header is a basic means of controlling caches; it tells all caches how long the associated representation is fresh for

//Cache-Control directives control who can cache the response, under which conditions, and for how long.
//The Cache-Control header was defined as part of the HTTP/1.1 specification and supersedes previous headers (for example, Expires) used to define response caching policies.
//no-store — instructs caches not to keep a copy of the representation under any conditions. It simply disallows intermediate caches from storing any version of the returned response—for example, one containing private personal or banking data. Every time the user requests this asset, a request is sent to the server and a full response is downloaded.

//no-cache — forces caches to submit the request to the origin server for validation before releasing a cached copy, every time. This is useful to assure that authentication is respected (in combination with public), or to maintain rigid freshness, without sacrificing all of the benefits of caching.
//"no-cache" indicates that the returned response can't be used to satisfy a subsequent request to the same URL without first checking with the server if the response has changed. As a result, if a proper validation token (ETag) is present, no-cache incurs a roundtrip to validate the cached response, but can eliminate the download if the resource has not changed.
//If you’d like such pages to be cacheable, but still authenticated for every user, combine the Cache-Control: public and no-cache headers. This tells the cache that it must submit the new client’s authentication information to the origin server before releasing the representation from the cache. This would look like:
//Cache-Control: public, no-cache

//public — marks authenticated responses as cacheable; normally, if HTTP authentication is required, responses are automatically private.
//If the response is marked as "public", then it can be cached, even if it has HTTP authentication associated with it, and even when the response status code isn't normally cacheable. Most of the time, "public" isn't necessary, because explicit caching information (like "max-age") indicates that the response is cacheable anyway.

//private — allows caches that are specific to one user (e.g., in a browser) to store the response; shared caches (e.g., in a proxy) may not.

//must-revalidate — tells caches that they must obey any freshness information you give them about a representation. HTTP allows caches to serve stale representations under special conditions; by specifying this header, you’re telling the cache that you want it to strictly follow your rules.

//proxy-revalidate — similar to must-revalidate, except that it only applies to proxy caches.

//max-age=[seconds] — specifies the maximum amount of time that a representation will be considered fresh. Similar to Expires, this directive is relative to the time of the request, rather than absolute. [seconds] is the number of seconds from the time of the request you wish the representation to be fresh for.

//s-maxage=[seconds] — similar to max-age, except that it only applies to shared (e.g., proxy) caches.

//When both Cache-Control and Expires are present, Cache-Control takes precedence.

//By default, pages protected with HTTP authentication are considered private; they will not be kept by shared caches. However, you can make authenticated pages public with a Cache-Control: public header;

//SSL pages are not cached (or decrypted) by proxy caches

func (mw cachemw) Download(url string) (output []byte, err error) {
	redisValue, err := redisCon.GetValue(url)
	if err == nil {
		var sResponse downloader.SplashResponse
		if err := json.Unmarshal(redisValue, &sResponse); err != nil {
			logger.Println("Json Unmarshall error", err)
		}
		//Error responses: a 404 (Not Found) may be cached.
		if sResponse.Obj.Status == 404 {
			return nil, fmt.Errorf("Error: 404. NOT FOUND")
		}
		output, err = sResponse.GetHTML()
		if err != nil {
			logger.Printf(err.Error())
		}
		return output, err
	}

	resp, err := mw.ParseService.GetResponse(url)
	//logger.Println(resp)
	//logger.Println(respErr)
	//
	if err == nil || resp.Obj.Status == 404 {
		response, err := json.Marshal(resp)
		if err != nil {
			logger.Printf(err.Error())
		}
		err = redisCon.SetValue(url, response)
		if err != nil {
			logger.Printf("%s: %s", errRedisSet, err.Error())
		}

	}
	if err != nil {
		return nil, err
	}
	output, err = resp.GetHTML()
	if err != nil {
		return nil, err
	}
	return
}

func (mw cachemw) ParseData(payload []byte) (output []byte, err error) {
	p, err := parser.NewParser(payload)
	if err != nil {
		return nil, err
	}
	rediskey := fmt.Sprintf("%s-%s", p.Format, p.PayloadMD5)
	redisValue, err := redisCon.GetValue(rediskey)
	if err == nil {
		return redisValue, nil
	}

	output, err = mw.ParseService.ParseData(payload)

	err = redisCon.SetValue(rediskey, output)
	if err != nil {
		logger.Printf("%s: %s", errRedisSet, err.Error())
	}
	return
}
