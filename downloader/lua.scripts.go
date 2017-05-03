package downloader

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
var baseLUA = `
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
  local responses = {}
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
  local ok, reason = splash:go(url)
  assert(splash:wait(%f))
  cookies = splash:get_cookies()
  if not ok then
       return {
        reason = reason,
     --   request = r.request.info,
     --   response = r.info,
      }
  end
  return {
      cookies = cookies,
      request = r.request.info,
      response = r.info,
	    html = splash:html(),       
  } 
end
`
//LUA script for formdata Post requests
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

var LUAPostFormData = `
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
  local responses = {}
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
  local ok, reason = splash:go{url,
    formdata={%s},
    http_method="POST"}
  assert(splash:wait(%f))
  cookies = splash:get_cookies()
  if not ok then
       return {
        reason = reason,
     --   request = r.request.info,
     --   response = r.info,
      }
  end
  return {
      cookies = cookies,
      request = r.request.info,
      response = r.info,
	    html = splash:html(),       
  } 
end
`

//LUA script which add cookies to request before sending it to browser
//It may be used for processing pages after initial authentication
//In the first step formdata with auth info is passed to a web page.
//Response object headers may contain an Object like
//name: "Set-Cookie"
//value: "session_id=29d7b97879209ca89316181ed14eb01f; path=/; httponly"
//These parameters should be passed to the next pages on the same domain.
//splash:add_cookie{"session_id", "29d7b97879209ca89316181ed14eb01f", "/", domain="example.com"}
    
var LUASetCookie = `
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
  local responses = {}
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
  splash:add_cookie{%s}
  local ok, reason = splash:go(url)
  assert(splash:wait(%f))
  if not ok then
       return {
        reason = reason,
     --   request = r.request.info,
     --   response = r.info,
      }
  end
  return {
      request = r.request.info,
      response = r.info,
	    html = splash:html(),       
  } 
end
`


var debugLUA = `
treat = require("treat")
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
  local urls = {}
	local resp_urls = {}
  splash:on_response(function (response)
  url = remove_trailing_slash(url)
  resp_url = remove_trailing_slash(response.info.url)
	table.insert(urls, url)
	table.insert(resp_urls, resp_url)
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
  local ok, reason = splash:go(url)
  assert(splash:wait(1))
  if not ok then
       return {
        reason = reason,
     --   request = r.request.info,
     --   response = r.info,
      }
  end
  return {
     -- request = r.request.info,
     -- response = r.info,
			urls = treat.as_array(urls),
			resp_urls = treat.as_array(resp_urls),
	    html = splash:html(),       
  } 
end
`