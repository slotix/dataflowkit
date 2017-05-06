package splash

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func paramsToLuaTable(params string) string {
	re := regexp.MustCompile("([\\w-]+)=([\\w%\\.]+)(&)?")
	p := re.ReplaceAllString(params, "$1=\"$2\",")
	return p
}

func (r *Response) setCookieToLUATable() (string, error) {
	headers := r.Response.Headers.(http.Header)
	setCookie := headers.Get("Set-Cookie")
	if setCookie != "" {
		cookies := r.Cookies
		for _, c := range cookies {
			logger.Println(c.Name, setCookie)
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
					strconv.FormatBool(c.HTTPOnly),
					strconv.FormatBool(c.Secure))
				return LUA, nil
			}
		}
	}
	return "", fmt.Errorf("No cookies in response")
}
