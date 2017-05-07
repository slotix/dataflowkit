package splash

import "time"

type Connection struct {
	Host            string
	User            string
	Password        string
	Timeout         int
	ResourceTimeout int
	LUAScript       string
}

type Request struct {
	URL     string  `json:"url"`
	Params  string  `json:"params,omitempty"` //params used for passing formdata to LUA script
	Cookies string  `json:"cookie,omitempty"` //add cookie to request
	Func    string  `json:"func,omitempty"`
	Wait    float64 `json:"wait,omitempty"` //Time in seconds to wait until js scripts loaded. Sometimes wait parameter should be set to more than default 0,5. It allows to finish js scripts execution on a web page.
}

type Cookies []struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Expires  time.Time `json:"expires,omitempty"`
	Domain   string    `json:"domain"`
	Secure   bool      `json:"secure"`
	Path     string    `json:"path"`
	HTTPOnly bool      `json:"httpOnly"`
}

type Response struct {
	HTML    string  `json:"html"`
	Reason  string  `json:"reason"`
	Cookies Cookies `json:"cookies"`
	Request struct {
		Cookies     Cookies `json:"cookies"`
		Method      string  `json:"method"`
		HeadersSize int     `json:"headersSize"`
		URL         string  `json:"url"`
		HTTPVersion string  `json:"httpVersion"`
		QueryString []struct {
			Value string `json:"value"`
			Name  string `json:"name"`
		} `json:"queryString"`
		//	Headers  []map[string]string `json:"headers"`
		//Headers  Headers `json:"headers"`
		Headers interface{} `json:"headers"`

		BodySize int `json:"bodySize"`
	} `json:"request"`
	Response struct {
		Headers interface{} `json:"headers"`
		//Headers    []map[string]string `json:"headers"`
		//Headers     Headers `json:"headers"`
		Cookies     Cookies `json:"cookies"`
		HeadersSize int     `json:"headersSize"`
		Ok          bool    `json:"ok"`
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
	} `json:"response"`
}

type PingResponse struct {
	Maxrss int    `json:"maxrss"`
	Status string `json:"status"`
}
