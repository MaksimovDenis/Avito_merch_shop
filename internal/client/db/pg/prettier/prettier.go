package prettier

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	PlaceholderDollar   = "$"
	PlaceholderQuestion = "?"
)

func Pretty(query string, placeholder string, args ...any) string {
	for idx, param := range args {
		var value string
		switch val := param.(type) {
		case string:
			value = fmt.Sprintf("%q", val)
		case []byte:
			value = fmt.Sprintf("%q", string(val))
		default:
			value = fmt.Sprintf("%v", val)
		}

		query = strings.Replace(query, fmt.Sprintf("%s%s", placeholder, strconv.Itoa(idx+1)), value, -1) // nolint
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.TrimSpace(query)
}
