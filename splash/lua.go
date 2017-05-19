package splash

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)
//paramsToLuaTable generates JSON string 
func paramsToLuaTable(params string) string {
	if params == "" {
		return params
	}
	re := regexp.MustCompile("([\\w-]+)=([\\w%\\.]+)(&)?")
	p := re.ReplaceAllString(params, "\"$1\":\"$2\",")
	p = strings.TrimSuffix(p,",") //remove last ","
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

func (r *Response) setCookieToLUATable() (string, error) {
	//	logger.Printf("%T - %s", r.Response.Headers, r.Response.Headers)
	headers := r.Response.Headers.(http.Header)
	setCookie := headers.Get("Set-Cookie")
	if setCookie != "" {
		cookies := r.Cookies
		for _, c := range cookies {
			//logger.Println(c.Name, setCookie)
			if strings.Contains(setCookie, c.Name) {
				//cookies = splash:add_cookie{name, value, path=nil, domain=nil, expires=nil, httpOnly=nil, secure=nil}
				//cookieLUA := `"session_id", "29d7b97879209ca89316181ed14eb01f", "/", domain="example.com"`
				expires := "nil"
				if !c.Expires.IsZero() {
					expires = c.Expires.String()
				}
				//logger.Println(c.Expires.IsZero())
				LUA := fmt.Sprintf(`"%s", "%s", path="%s", domain="%s", expires=%s, httpOnly=%s, secure=%s`,
					c.Name,
					c.Value,
					c.Path,
					c.Domain,
					expires,
					//	strconv.FormatBool(c.HTTPOnly),
					strconv.FormatBool(c.HttpOnly),
					strconv.FormatBool(c.Secure))
				return LUA, nil
			}
		}
	}
	return "", fmt.Errorf("No cookies in response")
}
