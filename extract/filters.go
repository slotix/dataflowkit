package extract

import (
	"strings"
)

func filterText(data string, filters []string) string{
	for _, filter := range filters {
		switch strings.ToLower(filter) {
		case "trim":
			data = strings.TrimSpace(data)
		case "lowercase":
			data = strings.ToLower(data)
		case "uppercase":
			data = strings.ToUpper(data)
		case "capitalize":
			data = strings.Title(data)			
		}
	}
	return data
}

func filterTextMW(data string, filters ...func(data string) string) string {
	for _, filter := range filters {
		data = filter(data)
	}
	return data
}
