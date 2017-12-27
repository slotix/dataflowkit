package splash

import (
	//"testing"
)

/* func TestGetResponse(t *testing.T) {
	//	url := "http://127.0.0.1:8050/execute?url=http%3A%2F%2Fgoogle.com&timeout=20&resource_timeout=30&lua_source=%0Afunction+string.ends%28String%2CEnd%29%0A++return+End%3D%3D%27%27+or+string.sub%28String%2C-string.len%28End%29%29%3D%3DEnd%0Aend%0Afunction+remove_trailing_slash%28text%29%0A++if+string.ends%28text%2C+%22%2F%22%29+then%0A++++text+%3D+text%3Asub%281%2C+-2%29%0A++end%0A++return+text%0Aend%0A%0Afunction+main%28splash%29%0A++local+url+%3D+splash.args.url%0A++local+responses+%3D+%7B%7D%0A++splash%3Aon_response%28function+%28response%29%0A++url+%3D+remove_trailing_slash%28url%29%0A++resp_url+%3D+remove_trailing_slash%28response.info.url%29%0A%09if+resp_url+%3D%3D+url+then%0A++++status+%3D+response.info.status%0A++++is_redirect+%3D+status+%3D%3D+301+or+status+%3D%3D+302%0A++++if+is_redirect+then%0A++++++url+%3D+response.info.redirectURL%0A++++elseif+status+%3D%3D+200+then%0A++%09%09%09r+%3D+response%0A++++end%0A++end%0A++end%29%0A++local+ok%2C+reason+%3D+splash%3Ago%28url%29%0A++assert%28splash%3Await%280.500000%29%29%0A++cookies+%3D+splash%3Aget_cookies%28%29%0A++if+not+ok+then%0A+++++++return+%7B%0A++++++++reason+%3D+reason%2C%0A+++++--+++request+%3D+r.request.info%2C%0A+++++--+++response+%3D+r.info%2C%0A++++++%7D%0A++end%0A++return+%7B%0A++++++cookies+%3D+cookies%2C%0A++++++request+%3D+r.request.info%2C%0A++++++response+%3D+r.info%2C%0A%09++++html+%3D+splash%3Ahtml%28%29%2C+++++++%0A++%7D+%0Aend%0A"

	url := "http://127.0.0.1:8050/execute?url=http%3A%2F%2Fdiesel.elcat.kg%2Findex.php%3Fs%3Defdce93be39d164b431291d301b7e368%26app%3Dcore%26module%3Dglobal%26section%3Dlogin%26do%3Dprocess&timeout=20&resource_timeout=30&lua_source=%0Afunction+string.ends%28String%2CEnd%29%0A++return+End%3D%3D%27%27+or+string.sub%28String%2C-string.len%28End%29%29%3D%3DEnd%0Aend%0Afunction+remove_trailing_slash%28text%29%0A++if+string.ends%28text%2C+%22%2F%22%29+then%0A++++text+%3D+text%3Asub%281%2C+-2%29%0A++end%0A++return+text%0Aend%0A%0Afunction+main%28splash%29%0A++local+url+%3D+splash.args.url%0A++local+responses+%3D+%7B%7D%0A++splash%3Aon_response%28function+%28response%29%0A++url+%3D+remove_trailing_slash%28url%29%0A++resp_url+%3D+remove_trailing_slash%28response.info.url%29%0A%09if+resp_url+%3D%3D+url+then%0A++++status+%3D+response.info.status%0A++++is_redirect+%3D+status+%3D%3D+301+or+status+%3D%3D+302%0A++++if+is_redirect+then%0A++++++url+%3D+response.info.redirectURL%0A++++elseif+status+%3D%3D+200+then%0A++%09%09%09r+%3D+response%0A++++end%0A++end%0A++end%29%0A++local+ok%2C+reason+%3D+splash%3Ago%7Burl%2C%0A++++formdata%3D%7Bauth_key%3D%22880ea6a14ea49e853634fbdc5015a024%22%2Creferer%3D%22http%253A%252F%252Fdiesel.elcat.kg%252F%22%2Cips_username%3D%22dm_%22%2Cips_password%3D%22dmsoft%22%2CrememberMe%3D%221%22%2C%7D%2C%0A++++http_method%3D%22POST%22%7D%0A++assert%28splash%3Await%280.500000%29%29%0A++cookies+%3D+splash%3Aget_cookies%28%29%0A++if+not+ok+then%0A+++++++return+%7B%0A++++++++reason+%3D+reason%2C%0A+++++--+++request+%3D+r.request.info%2C%0A+++++--+++response+%3D+r.info%2C%0A++++++%7D%0A++end%0A++return+%7B%0A++++++cookies+%3D+cookies%2C%0A++++++request+%3D+r.request.info%2C%0A++++++response+%3D+r.info%2C%0A%09++++html+%3D+splash%3Ahtml%28%29%2C+++++++%0A++%7D+%0Aend%0A"
	req := Request{URL: url}
	response, err := req.GetResponse()
	if err != nil {
		logger.Println(err)
	}
	//	response.Response.Headers = castHeaders(response.Response.Headers)

	//logger.Printf("%T - %s", response.Response.Headers, response.Response.Headers)
	logger.Printf("%T - %v- %d\n", response.Request.Cookies, response.Request.Cookies, len(response.Request.Cookies))
	logger.Printf("%T - %v\n", response.Request.URL, response.Request.URL)
} */
