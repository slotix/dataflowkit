package splash

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParamsToLUATable(t *testing.T) {
	params := "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=password&rememberMe=1"
	p := paramsToLuaTable(params)
	assert.Equal(t, `{"auth_key":"880ea6a14ea49e853634fbdc5015a024","referer":"http%3A%2F%2Fexample.com%2F","ips_username":"user","ips_password":"password","rememberMe":"1"}`, p) 

}

func TestGenerateCookie(t *testing.T) {
	cookie := `example_uzt=72e3502635d3af8fa2916cf397e93fee; expires=Tue, 04-Jul-2017 13:28:36 GMT; Max-Age=2592000; path=/; domain=.example.com; HttpOnly
heureka_s=1; expires=Mon, 04-Jun-2018 13:28:36 GMT; Max-Age=31536000; path=/; domain=.example.com`
	s, err := generateCookie(cookie)
	if err != nil {
		logger.Println(nil)
	}
	assert.Equal(t, `[{"name":"example_uzt", "value":"72e3502635d3af8fa2916cf397e93fee", "path":"/", "domain":".example.com", "expires":"Tue, 04-Jul-2017 13:28:36 GMT", "httpOnly":true, "secure":false},{"name":"heureka_s", "value":"1", "path":"/", "domain":".example.com", "expires":"Mon, 04-Jun-2018 13:28:36 GMT", "httpOnly":false, "secure":false}]`, s)
}

/*
func TestParamsToJSON(t *testing.T) {
	//params := "auth_key=test&referer=http%3A%2F%2Fexample.com%2F&ips_username=test&ips_password=test&rememberMe=1"
	params := "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fdiesel.elcat.kg%2F&ips_username=dm_&ips_password=dmsoft&rememberMe=1"
//	logger.Println(paramsToJSON(params))
	assert.Equal(t, 
	`{"auth_key":["880ea6a14ea49e853634fbdc5015a024"],"ips_password":["dmsoft"],"ips_username":["dm_"],"referer":["http://diesel.elcat.kg/"],"rememberMe":["1"]}`, paramsToJSON(params))	

}
*/