package splash

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

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Cookie struct {
	Name  string
	Value string

	Path       string // optional
	Domain     string // optional
	Expires    string // optional
	RawExpires string // for reading cookies only
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}

//SResponse returned by splash as a part of Response
//It is needed to be passed to caching middleware
type SResponse struct {
	Headers interface{} `json:"headers"`
	//Headers     []Header      `json:"headers"`
	HeadersSize int      `json:"headersSize"`
	Cookies     []Cookie `json:"cookies"`
	//Cookies []Cookie `json:"cookies"`
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

//SRequest returned by splash as a part of Response
//It is needed to be passed to caching middleware
type SRequest struct {
	Method  string      `json:"method"`
	Headers interface{} `json:"headers"`
	//Headers     []Header      `json:"headers"`
	HeadersSize int      `json:"headersSize"`
	Cookies     []Cookie `json:"cookies"`
	//Cookies     []Cookie `json:"cookies"`
	URL         string `json:"url"`
	HTTPVersion string `json:"httpVersion"`
	QueryString []struct {
		Value string `json:"value"`
		Name  string `json:"name"`
	} `json:"queryString"`
	BodySize int `json:"bodySize"`
}

//Response returned by splash
//It includes html body, response, request
type Response struct {
	URL    string `json:"url"`
	HTML   string `json:"html"`
	Reason string `json:"reason"`
	//Cookies             []Cookie   `json:"cookies"`
	Request             *SRequest  `json:"request"`
	Response            *SResponse `json:"response"`
	Cacheable           bool
	CacheExpirationTime int64
}

type PingResponse struct {
	Maxrss int    `json:"maxrss"`
	Status string `json:"status"`
}

type gcResponse struct {
	CachedArgsRemoved  int    `json:"cached_args_removed"`
	PyObjectsCollected int    `json:"pyobjects_collected"`
	Status             string `json:"status"`
}
