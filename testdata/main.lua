--treat = require("treat")
--function string.starts(String, Start)
-- return string.sub(String,1,string.len(Start))==Start
--end

function string.ends(String,End)
   return End=='' or string.sub(String,-string.len(End))==End
end

function split(str, patt)
	vals = {}; valindex = 0; word = ""
	-- need to add a trailing separator to catch the last value.
	str = str .. patt
	for i = 1, string.len(str) do
	
		cha = string.sub(str, i, i)
		if cha ~= patt then
			word = word .. cha
		else
			if word ~= nil then
				vals[valindex] = word
				valindex = valindex + 1
				word = ""
			else
				-- in case we get a line with no data.
				break
			end
		end 
		
	end	
	return vals
end

-----------------------------------------------------------------------------  
-- Modified URI parsing function from LuaSocket toolkit. 
-- Author: Diego Nehab.
-- https://github.com/diegonehab/luasocket/blob/master/src/url.lua
-- Modifications have been made due to restrictions of sandbox 
-- described at https://github.com/scrapinghub/splash/blob/2.3.x/splash/lua_modules/sandbox.lua
-- test1 = "http://www.example.com/cgilua/index.lua?a=2#there"
-- test2 = "ftp://root:passwd@unsafe.org/pub/virus.exe;type=i"
-----------------------------------------------------------------------------


function parse(url)
  local parsed = {}
  --fragment
  i,j = string.find(url, "#(.*)$")
  if i~=nil then
    parsed.fragment = string.sub(url, i+1,j)
    url = string.sub(url,1,i-1)
  end

  --scheme
  i,j = string.find(url, "^([%w][%w%+%-%.]*)%:")
  if i~=nil then
    parsed.scheme = string.sub(url, i,j-1)
    url = string.sub(url,j+1)
  end
  --authority

  i,j = string.find(url, "^//([^/]*)")
  if i~=nil then
    parsed.authority = string.sub(url, i+2,j)
    url = string.sub(url,j+2)
  end

  --query
  i,j = string.find(url, "%?(.*)")
  if i~=nil then
    parsed.query = string.sub(url, i+1,j)
    url = string.sub(url,1,i-1)
  end

  --params
  i,j = string.find(url, "%;(.*)")
  if i~=nil then
    parsed.params = string.sub(url, i+1,j)
    url = string.sub(url,1,i-1)
  end

  -- path is whatever was left
  if url ~= "" then parsed.path = url end
  local authority = parsed.authority
  if not authority then return parsed end

  --userinfo
  i,j = string.find(authority,"^([^@]*)@")
  if i~=nil then
    parsed.userinfo = string.sub(authority, i,j-1)
    authority = string.sub(authority,j+1)
  end

  --port
  i,j = string.find(authority,":([^:%]]*)$")
  if i~=nil then
    parsed.port = string.sub(authority, i+1)
    authority = string.sub(authority,1,i-1)
  end
  if authority ~= "" then
    -- IPv6?
    --parsed.host = string.match(authority, "^%[(.+)%]$") or authority
    parsed.host = authority
  end
  local userinfo = parsed.userinfo
  if not userinfo then return parsed end
  --user/password
  i,j = string.find(userinfo, ":([^:]*)$")
  parsed.password = string.sub(userinfo, i+1, j)
  parsed.user = string.sub(userinfo, 1, i-1)
  print(parsed.password)
  print(parsed.user)
  return parsed
end
local url = "http://www.example.com/cgilua/index.lua?a=2#there"
url1 = "http://google.com"
url2 = "http://google.com/"

i,j = string.find(url, "#(.*)$")
  if i~=nil then
--e = string.ends(url,"/")
--print(e)
--local parsed_host = parse(url).host
--s = split(parsed_host,".")
--print (s)

function main1(splash)
  local responses = {}
  --local parsed_hosts = {}
  local parsed = {}
  local url = splash.args.url
  local initial_url = url
  --local initial_host = parse(url).host
  splash.response_body_enabled = true
  splash:on_response(function (response)
    status = response.info.status
    if status == 200 and response.info.url == url then
      current_url_host = parse(url).host
      table.insert(responses, response.info)
      --parsed = parse(url).host
      --table.insert(parsed_hosts, parsed)
    elseif response.info.url == initial_url and (status == 301 or status == 302)  then
      url = response.info.redirectURL
    end
  end)
  assert(splash:go(url))
  assert(splash:wait(1))

  return {
    result = treat.as_array(responses),
 --   hosts = treat.as_array(parsed_hosts),
    --initial_host = initial_host,
    --current_url_host = current_url_host,
    --parsed = treat.as_array(parsed),
    --parsed_urls = treat.as_array(parsed_urls),
  }
end