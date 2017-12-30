package splash

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
			}
			cf := strings.Split(g, "=")
			switch strings.ToLower(strings.Trim(cf[0], " ")) {
			case "expires":
				cookie.Expires = cf[1]
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

	return fmt.Sprintf("[%s]", strings.Join(out, ",")), nil
}

func (r *Response) getSetCookie() (string, error) {
	headers := r.Response.Headers.(http.Header)
	setCookie := headers.Get("Set-Cookie")
	if setCookie == "" {
		return "", fmt.Errorf("No Set-Cookie found")
	}
	return setCookie, nil
}

//  SetCookieToNextRequest retrieves Set-Cookie from current Request to send it further to the next request withing the same domain.
func (r *Response) SetCookieToNextRequest(req *Request) error {
	setCookie, err := r.getSetCookie()
	if err != nil {
		return err
	}
	//it may be more than one cookie in Set-Cookie
	//heu_uzt=72e3502635d3af8fa2916cf397e93fee; expires=Tue, 04-Jul-2017 13:28:36 GMT; Max-Age=2592000; path=/; domain=.heu.tt
	//heu_s=1; expires=Mon, 04-Jun-2018 13:28:36 GMT; Max-Age=31536000; path=/; domain=.heu.tt

	//cookies = splash:add_cookie{name, value, path=nil, domain=nil, expires=nil, httpOnly=nil, secure=nil}
	//cookieLUA := `"session_id", "29d7b97879209ca89316181ed14eb01f", "/", domain="example.com"`
	cookie, err := generateCookie(setCookie)
	if err != nil {
		logger.Error(err)
	}
	req.Cookies = cookie
	return nil
}



