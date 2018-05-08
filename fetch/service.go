package fetch

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/juju/persistent-cookiejar"
	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/storage"
	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
)

// Service defines Fetch service interface
type Service interface {
	Response(req FetchRequester) (FetchResponser, error)
	Fetch(req FetchRequester) (io.ReadCloser, error)
}

// FetchService implements service with empty struct
type FetchService struct {
}

// ServiceMiddleware defines a middleware for a Fetch service
type ServiceMiddleware func(Service) Service

//Response returns splash.Response
func (fs FetchService) Response(req FetchRequester) (FetchResponser, error) {
	var fetcher Fetcher
	switch req.(type) {
	case BaseFetcherRequest:
		fetcher = NewFetcher(Base)
	case splash.Request:
		fetcher = NewFetcher(Splash)
	}
	var (
		jar     *cookiejar.Jar
		cookies []byte
		cArr    []*http.Cookie
		s       storage.Store
	)
	if req.GetUserToken() != "" {
		storageType, err := storage.TypeString(viper.GetString("STORAGE_TYPE"))
		if err != nil {
			return nil, err
		}
		s = storage.NewStore(storageType)
		cookies, _ = s.Read(req.GetUserToken())
		jarOpts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
		jar, err = cookiejar.New(jarOpts)
		if err != nil {
			logger.Error("Failed to create Cookie Jar")

		}
		cArr = []*http.Cookie{}
		if len(cookies) != 0 {
			err = json.Unmarshal(cookies, &cArr)
			if err != nil {
				return nil, err
			}
			u, err := url.Parse(req.GetURL())
			if err != nil {
				return nil, err
			}
			tempCarr := []*http.Cookie{}
			for i := 0; i < len(cArr); i++ {
				c := cArr[i]
				if u.Host == c.Domain {
					tempCarr = append(tempCarr, c)
					cArr = append(cArr[:i], cArr[i+1:]...)
					i--
				}
			}
			jar.SetCookies(u, tempCarr)
		}
		fetcher.SetCookieJar(jar)
	}
	//res, err := fetcher.Fetch(req)
	res, err := fetcher.Response(req)
	if err != nil {
		return nil, err
	}
	if req.GetUserToken() != "" {
		jar = fetcher.GetCookieJar()
		cArr = append(cArr, jar.AllCookies()...)
		cookies, err = json.Marshal(cArr)
		//logger.Info(string(cookies))
		if err != nil {
			return nil, err
		}
		err = s.Write(req.GetUserToken(), cookies, 0)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

//Fetch downloads web page content and returns it
func (fs FetchService) Fetch(req FetchRequester) (io.ReadCloser, error) {
	res, err := fs.Response(req)
	if err != nil {
		return nil, err
	}
	return res.GetHTML()
}
