package fetch

func RobotsTxtMiddleware() ServiceMiddleware {
	return func(next Service) Service {
		return robotstxtMiddleware{next}
	}
}

type robotstxtMiddleware struct {
	Service
}

//Fetches response from req.URL, then pass response.URL to Robots.txt validator.
//issue #1 https://github.com/slotix/dataflowkit/issues/1
func (mw robotstxtMiddleware) Fetch(req FetchRequester) (response FetchResponser, err error) {
	url := req.GetURL()
	//to avoid recursion while retrieving robots.txt
	if !IsRobotsTxt(url) {
		_, err := RobotstxtData(url)
		if err != nil {
			return nil, err
		}
	}
	response, err = mw.Service.Fetch(req)
	if err != nil {
		return nil, err
	}
	return response, err
}

// func (mw robotstxtMiddleware) Fetch(req interface{}) (output interface{}, err error) {

// 	output, err = mw.Service.Fetch(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	sResponse := output.(*splash.Response)
// 	url := sResponse.URL

// 	if !splash.IsRobotsTxt(url) {
// 		_, err := splash.RobotstxtData(url)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return output, err
// }
