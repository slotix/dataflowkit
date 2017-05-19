package splash

import "testing"

func TestOut(t *testing.T) {
	params := "auth_key=test&referer=http%3A%2F%2Fexample.com%2F&ips_username=test&ips_password=test&rememberMe=1"
	logger.Println(paramsToLuaTable(params))

}

func TestParamsToJSON(t *testing.T) {
	//params := "auth_key=test&referer=http%3A%2F%2Fexample.com%2F&ips_username=test&ips_password=test&rememberMe=1"
	params := "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fdiesel.elcat.kg%2F&ips_username=dm_&ips_password=dmsoft&rememberMe=1"
	logger.Println(paramsToJSON(params))
}