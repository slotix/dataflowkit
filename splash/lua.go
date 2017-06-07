package splash

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)


func generateCookie(setCookie string) (string, error) {
	//split strings in case of more than one cookie
	ss := strings.Split(setCookie, "\n")
	//split groups divided by ;

	var out []string
	for _, s := range ss {
		cookie := Cookie{}
		groups := strings.Split(s, ";")
		//split cookie fields divided by "="
		for i, g := range groups {
			if i == 0 {
				cf := strings.Split(g, "=")
				cookie.Name = cf[0]
				cookie.Value = cf[1]
				//	logger.Println(groups)
			}
			cf := strings.Split(g, "=")
		//	logger.Println(cf[0])
			switch strings.ToLower(strings.Trim(cf[0], " ")) {
			case "expires":
				cookie.Expires = cf[1]
		//		logger.Println(cookie)
			//case "Max-Age":
			//cookie.MaxAge = cf[1]
			case "path":
				cookie.Path = cf[1]
			case "domain":
				cookie.Domain = cf[1]
			case "httponly":
				cookie.HttpOnly = true
			case "secure":
				cookie.Secure = true
			}
		}
		//LUA := fmt.Sprintf(`name="%s", value="%s", path="%s", domain="%s", expires="%s", httpOnly=%s, secure=%s`,
		LUA := fmt.Sprintf(`{"name":"%s", "value":"%s", "path":"%s", "domain":"%s", "expires":"%s", "httpOnly":%s, "secure":%s}`,
			cookie.Name,
			cookie.Value,
			cookie.Path,
			cookie.Domain,
			cookie.Expires,
			strconv.FormatBool(cookie.HttpOnly),
			strconv.FormatBool(cookie.Secure))
		out = append(out, LUA)
	}

	return fmt.Sprintf("[%s]",strings.Join(out,",")), nil
}

func (r *Response) setCookieToLUATable() (string, error) {
	headers := r.Response.Headers.(http.Header)
	setCookie := headers.Get("Set-Cookie")
	if setCookie == "" {
		return "", fmt.Errorf("No Set-Cookie found")
	}
	//it may be more than one cookie in Set-Cookie
	//heu_uzt=72e3502635d3af8fa2916cf397e93fee; expires=Tue, 04-Jul-2017 13:28:36 GMT; Max-Age=2592000; path=/; domain=.heu.tt
	//heu_s=1; expires=Mon, 04-Jun-2018 13:28:36 GMT; Max-Age=31536000; path=/; domain=.heu.tt

	//cookies = splash:add_cookie{name, value, path=nil, domain=nil, expires=nil, httpOnly=nil, secure=nil}
	//cookieLUA := `"session_id", "29d7b97879209ca89316181ed14eb01f", "/", domain="example.com"`
	cookie, err := generateCookie(setCookie)
	if err != nil{
		logger.Println(err)
	}

	return cookie, nil
}

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
