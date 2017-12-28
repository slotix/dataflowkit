package splash

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

//paramsToLuaTable generates JSON string
func paramsToLuaTable(params string) string {
	if params == "" {
		return params
	}
	re := regexp.MustCompile("([\\w-]+)=([\\w%\\.]+)(&)?")
	p := re.ReplaceAllString(params, "\"$1\":\"$2\",")
	p = strings.TrimSuffix(p, ",") //remove last ","
	p = fmt.Sprintf("{%s}", p)
	return p
}

func paramsToJSON(params string) string {
	if params == "" {
		return params
	}
	m, err := url.ParseQuery(params)
	if err != nil {
		logger.Error(err)
		return ""
	}
	jsonString, err := json.Marshal(m)
	if err != nil {
		logger.Error(err)
		return ""
	}
	return string(jsonString)
}

func toTime(s string) (t time.Time, err error) {
	// Get rid of the quotes "" around the value.
	// A second option would be to include them
	// in the date format string instead, like so below:
	//   time.Parse(`"`+time.RFC3339Nano+`"`, s)

	//s = s[1 : len(s)-1]

	t, err = time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05.999999999Z0700", s)
		if err != nil {
			return time.Time{}, err
		}
	}

	return t, nil
}