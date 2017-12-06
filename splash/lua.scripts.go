package splash

//LUA script for robots.txt files processing
var robotsLUA = `
function main(splash) 
  local url = splash.args.url 
  local response = splash:http_get(url)
  return { 
    request = response.request.info,
    response = response.info,
  } 
end
`

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
function main(splash)
  cookies = splash.args.cookies -- set to "" when running the script in browserl
  formdata = splash.args.formdata -- set to "" when running the script in browser
  http_method = splash.args.http_method
  decoded = nil
  http_method = "GET"
  if formdata ~= "" then
    decoded = json.decode(formdata)
    http_method = "POST" 
  end
  if cookies ~= "" then
    cookies_array = json.decode(cookies)
    for k,v in next,cookies_array,nil
    do
  	  splash:add_cookie(v)
    end
  end

  local ok, reason = splash:go{
    splash.args.url,
    headers = splash.args.headers,
    http_method = http_method,
    formdata = decoded,
    body = splash.args.body,
    }
  assert(splash:wait(splash.args.wait))

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
  return {
    url = splash:url(),
    request = request,
    response = response,
    -- cookies = splash:get_cookies(),
    html = splash:html(),
  }
end
`




/*
//LUA script for general pages processing
var baseLUAOld = `
json = require("json")
function string.ends(String,End)
  return End=='' or string.sub(String,-string.len(End))==End
end
function remove_trailing_slash(text)
  if string.ends(text, "/") then
    text = text:sub(1, -2)
  end
  return text
end

function main(splash)
  local url = splash.args.url
  splash:init_cookies(splash.args.cookies)
  formdata = splash.args.formdata
  splash:on_response(function (response)
  url = remove_trailing_slash(url)
  resp_url = remove_trailing_slash(response.info.url)
	if resp_url == url then
    status = response.info.status
    is_redirect = status == 301 or status == 302
    if is_redirect then
      url = response.info.redirectURL
    elseif status == 200 then
  			r = response
    end
  end
  end)
  if formdata == "" then
      ok, reason = splash:go(url)
  else
      decoded = json.decode(formdata)
      ok, reason = splash:go{url,
      formdata= decoded,
      http_method="POST"}
  end
  assert(splash:wait(splash.args.wait))
  if not ok then
       return {
        reason = reason,
      }
  end
  return {
      cookies = splash:get_cookies(),
      request = r.request.info,
      response = r.info,
	    html = splash:html(),       
  } 
end
`
*/