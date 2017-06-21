json = require("json")
function main(splash)
  --cookies = '{"name":"heureka_uztt", "value":"72e3502635d3af8fa2916cf397e93fee", "path":"/", "domain":".heureka.sk", "expires":"Tue, 04-Jul-2017 13:28:36 GMT", "httpOnly":true, "secure":false}'
  --cookies = '{"name":"heureka_uztt", "value":"72e3502635d3af8fa2916cf397e93fee", "path":"/", "domain":".heureka.sk", "expires":"Tue, 04-Jul-2017 13:28:36 GMT", "httpOnly":true, "secure":false}'
  --splash:add_cookie(json.decode(cookies))
  --cookies = '{"name":"heureka", "value":"72e3502635d3af8fa2916cf397e93fee", "path":"/", "domain":".heureka.sk", "expires":"Tue, 04-Jul-2017 13:28:36 GMT", "httpOnly":true, "secure":false}'
  cookiess = '[{"name":"heureka111", "value":"72e3502635d3af8fa2916cf397e93fee", "path":"/", "domain":".heureka.sk", "expires":"Tue, 04-Jul-2017 13:28:36 GMT", "httpOnly":true, "secure":false},{"name":"huereka", "value":"72e3502635d3af8fa2916cf397e93fee", "path":"/", "domain":".heureka.sk", "expires":"Tue, 04-Jul-2017 13:28:36 GMT", "httpOnly":true, "secure":false}]'
  cookies_array = json.decode(cookiess)
  --for key,value in next,t,nil
  for k,v in next,cookies_array,nil
  do
  	splash:add_cookie(v)
  end
  
  formdata = ""
  http_method = splash.args.http_method
  decoded = nil
  http_method = "GET"
  if formdata ~= "" then
    decoded = json.decode(formdata) 
    http_method = "POST" 
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
        reason = reason,
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
    --cookies = splash:get_cookies(),
    cookies = cookies_array,
    html = splash:html(),
  }
end