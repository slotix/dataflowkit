package splash

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateCookie(t *testing.T) {
	cookie := `example_uzt=72e3502635d3af8fa2916cf397e93fee; expires=Tue, 04-Jul-2017 13:28:36 GMT; Max-Age=2592000; path=/; domain=.example.com; HttpOnly
heureka_s=1; expires=Mon, 04-Jun-2018 13:28:36 GMT; Max-Age=31536000; path=/; domain=.example.com`
	s, err := generateCookie(cookie)
	if err != nil {
		logger.Println(nil)
	}
	assert.Equal(t, `[{"name":"example_uzt", "value":"72e3502635d3af8fa2916cf397e93fee", "path":"/", "domain":".example.com", "expires":"Tue, 04-Jul-2017 13:28:36 GMT", "httpOnly":true, "secure":false},{"name":"heureka_s", "value":"1", "path":"/", "domain":".example.com", "expires":"Mon, 04-Jun-2018 13:28:36 GMT", "httpOnly":false, "secure":false}]`, s)
}


func TestGetSetCookie(t *testing.T) {
	
	type args struct {
		headers http.Header
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{	//there may be more than one cookie in Set-Cookie. 
			//But the only the first value will be returned with the given key.
			// More info: http/header.go func (h Header) Get(key string) string 
			name: "1",
			args: args{
				headers: http.Header{
					"Set-Cookie": []string{"heureka_uzt=e2728538559536f2bf3a99a90da3066d; expires=Thu, 08-Feb-2018 00:04:21 GMT; Secure; Max-Age=2592000; path=/; domain=.heureka.sk","heureka_s=1; expires=Wed, 09-Jan-2019 00:04:21 GMT; Max-Age=31536000; path=/; domain=.heureka.sk"},
					"Content-Type": []string {"text/html; charset=utf-8",},
					"Cache-Control": []string {"no-cache, no-store, must-revalidate"},
				},
			},
			want: `[{"name":"heureka_uzt", "value":"e2728538559536f2bf3a99a90da3066d", "path":"/", "domain":".heureka.sk", "expires":"Thu, 08-Feb-2018 00:04:21 GMT", "httpOnly":false, "secure":true}]`,
		},
		{	
			name: "Empty",
			args: args{
				headers: http.Header{
					"Set-Cookie": []string{""},
					"Content-Type": []string {"text/html; charset=utf-8",},
					"Cache-Control": []string {"no-cache, no-store, must-revalidate"},
				},
			},
			want: "",
		},
	
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSetCookie(tt.args.headers); got != tt.want {
				t.Errorf("GetSetCookie() = %v, want %v", got, tt.want)
			}
		})
	}
}
