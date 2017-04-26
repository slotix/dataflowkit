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
var generalLUA = `
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
  assert(splash:wait(%d))
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
//LUA script which process page which requires authentication
var generalLUAWithAuth = `
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
  assert(splash:wait(%d))
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