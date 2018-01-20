package fetch

import "github.com/slotix/dataflowkit/errs"

//RobotsTxtMiddleware checks if scraping of specified resource is allowed by robots.txt
func RobotsTxtMiddleware() ServiceMiddleware {
	return func(next Service) Service {
		return robotstxtMiddleware{next}
	}
}

type robotstxtMiddleware struct {
	Service
}

//Fetch gets response from req.URL, then passes response.URL to Robots.txt validator.
//issue #1 https://github.com/slotix/dataflowkit/issues/1
func (mw robotstxtMiddleware) Fetch(req FetchRequester) (response FetchResponser, err error) {
	url := req.GetURL()
	//to avoid recursion while retrieving robots.txt
	if !IsRobotsTxt(url) {
		robotsData, err := RobotstxtData(url)
		if err != nil {
			return nil, err
		}
		if !AllowedByRobots(url, robotsData) {
			//no need a body retrieve to get information about redirects
			r := BaseFetcherRequest{URL: url, Method: "HEAD"}
			resp, err := fetchRobots(r)
			if err != nil {
				return nil, err
			}
			//if initial URL is not equal to final URL (Redirected) f.e. domains are different
			//then try to fetch robots following by final URL
			finalURL := resp.GetURL()
			//	finalURL := resp.Request.URL.String()
			if url != finalURL {
				robotsData, err = RobotstxtData(finalURL)
				if err != nil {
					return nil, err
				}
				if !AllowedByRobots(finalURL, robotsData) {
					return nil, &errs.ForbiddenByRobots{finalURL}
				}
			} else {
				return nil, &errs.ForbiddenByRobots{url}
			}
			//	bfReq := BaseFetcherRequest{URL: finalURL}
			//	req = bfReq
		}

	}
	response, err = mw.Service.Fetch(req)
	if err != nil {
		return nil, err
	}
	return response, err
}
