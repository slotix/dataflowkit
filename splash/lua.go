package splash

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

//paramsToLuaTable generates JSON string
func paramsToLuaTable(params string) string {
	if params == "" {
		return params
	}
	re := regexp.MustCompile("([\\w-]+)=([\\w%\\.]+)(&)?")
	p := re.ReplaceAllString(params, "\"$1\":\"$2\",")
	p = strings.TrimSuffix(p, ",") //remove last ","
	p = fmt.Sprintf("{%s}", p)
	return p
}

func paramsToJSON(params string) string {
	if params == "" {
		return params
	}
	m, err := url.ParseQuery(params)
	if err != nil {
		logger.Println(err)
		return ""
	}
	jsonString, err := json.Marshal(m)
	if err != nil {
		logger.Println(err)
		return ""
	}
	return string(jsonString)
}

func toTime(s string) (t time.Time, err error) {
	//logger.Println(s)
	// Get rid of the quotes "" around the value.
	// A second option would be to include them
	// in the date format string instead, like so below:
	//   time.Parse(`"`+time.RFC3339Nano+`"`, s)

	//s = s[1 : len(s)-1]

	t, err = time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05.999999999Z0700", s)
		if err != nil {
			return time.Time{}, err
		}
	}

	return t, nil
}

/*
func (r *Response) setCookieToLUATable() (string, error) {
	//	logger.Printf("%T - %s", r.Response.Headers, r.Response.Headers)
	headers := r.Response.Headers.(http.Header)
	//logger.Printf("%T: %s",r.Response.Headers,r.Response.Headers)
	//headers := r.Response.castHeaders()
	setCookie := headers.Get("Set-Cookie")
	//logger.Printf(setCookie)
	if setCookie != "" {
		cookies := r.Cookies
		//logger.Printf("%T: %s",cookies,cookies)
		for _, c := range cookies {
			logger.Printf("%s::: %s", setCookie, c.Name)
			if strings.Contains(setCookie, c.Name) {
				//cookies = splash:add_cookie{name, value, path=nil, domain=nil, expires=nil, httpOnly=nil, secure=nil}
				//cookieLUA := `"session_id", "29d7b97879209ca89316181ed14eb01f", "/", domain="example.com"`
				expires := "nil"
				t, err := toTime(c.Expires.(string))
				if err != nil {
					return "", err
				}
				if !t.IsZero() {
					expires = c.Expires.(string)
				}

				LUA := fmt.Sprintf(`"%s", "%s", path="%s", domain="%s", expires=%s, httpOnly=%s, secure=%s`,
					c.Name,
					c.Value,
					c.Path,
					c.Domain,
					expires,
					//	strconv.FormatBool(c.HTTPOnly),
					strconv.FormatBool(c.HttpOnly),
					strconv.FormatBool(c.Secure))
				logger.Println(LUA)
				return LUA, nil
			}
		}
	}
	return "", fmt.Errorf("No Set-Cookie found")
}
*/
