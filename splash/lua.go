package splash

import (
	"fmt"
	"strings"
)

//LUA script for general pages processing
//Also formdata parameters may be passed
//For example it may be used for processing pages which require authentication
//formdata Example:
//local ok, reason = splash:go{url,
//    formdata={
//      auth_key = "880ea6a14ea49e853634fbdc5015a024",
//      ips_username = "user",
//      ips_password = "userpass",
//      rememberMe = "1",
//    },
// http_method="POST"}

//------

//Passing Cookies to request before sending it to browser
//It may be used for processing pages after initial authentication
//In the first step formdata with auth info is passed to a web page.
//Response object headers may contain an Object like
//name: "Set-Cookie"
//value: "session_id=29d7b97879209ca89316181ed14eb01f; path=/; httponly"
//This cookie should be passed to the next pages on the same domain.
//splash:add_cookie{"session_id", "29d7b97879209ca89316181ed14eb01f", "/", domain="example.com"}

var baseLUA = `
json = require("json")
function scroll2bottom(waitInterval, splash)
  local docHeight = 0
  local intervalFound = false
  while docHeight < splash:evaljs([[window.document.body.scrollHeight;]]) do
    docHeight = splash:evaljs([[window.document.body.scrollHeight]])
    local js = string.format(
        [[window.scrollTo(0, %s);]],
        tonumber(docHeight)
    )
    splash:runjs(js)
    splash:wait(waitInterval)
    if waitInterval < 5 and docHeight == splash:evaljs([[window.document.body.scrollHeight]]) then
      return intervalFound
    end
    if waitInterval < 5 and docHeight ~= splash:evaljs([[window.document.body.scrollHeight]]) then
      intervalFound = true
    end
    if waitInterval == 5 then
      return true
    end
  end
  return true
end

function main(splash, args)
  cookies = ""
  formdata = ""--args.formdata
headers = nil
  decoded = nil
  http_method = "GET"
  if formdata ~= "" then
    decoded = json.decode(formdata)
    http_method = "POST" 
  end
  if cookies ~= "" then
    cookies_array = json.decode(cookies)
    splash:init_cookies(cookies_array)
  end
if args.headers ~= "" then
	--headers = json.decode(args.headers)
end

  local ok, reason = splash:go{
    args.url,
    headers = headers,
    http_method = http_method,
    formdata = decoded,
    body = args.body,
    }
  assert(splash:wait(args.wait))

  local entries = splash:history()
  local last_entry = entries[#entries]
  if not ok then
       return {
        error = reason,
      }
  end
  if #entries>0  then
    	request = last_entry.request
    	response = last_entry.response
    end
  if args.scroll2bottom == "true" then
    waitInterval = 0.5
    while scroll2bottom(waitInterval, splash)==false do
      waitInterval = waitInterval + 0.5
    end
  end
  return {
    url = splash:url(),
    request = request,
    response = response,
    cookies = splash:get_cookies(),
    html = splash:html(),
    png = splash:png{width=640},
    har = splash:har(),
    waitInterval = waitInterval,
  }
end
`

func paramsToLuaTable(params string) string {
	//"auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=usr&ips_password=passw&rememberMe=0"
	if len(params) == 0 {
		return ""
	}
	formData := ""
	//formData := make(map[string])
	pairs := strings.Split(params, "&")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		formData += `"` + kv[0] + `":"` + kv[1] + `",`
	}
	formData = strings.TrimSuffix(formData, ",") //remove last ","
	formData = fmt.Sprintf("{%s}", formData)
	return formData
}
