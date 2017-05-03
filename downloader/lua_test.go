package downloader

import "testing"

func TestOut(t *testing.T) {
	params := "auth_key=test&referer=http%3A%2F%2Fexample.com%2F&ips_username=test&ips_password=test&rememberMe=1"
	logger.Println(paramsToLuaTable(params))

}
