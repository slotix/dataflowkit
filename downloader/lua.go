package downloader

import (
	"regexp"
)

func paramsToLuaTable(params string) string {
	re := regexp.MustCompile("([\\w-]+)=([\\w%\\.]+)(&)?")
	p := re.ReplaceAllString(params, "$1=\"$2\",")
	return p
}
