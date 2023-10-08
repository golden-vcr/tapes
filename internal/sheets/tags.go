package sheets

import "strings"

func parseTagHeading(s string) string {
	questionMarkPos := strings.Index(s, "?")
	if questionMarkPos >= 0 && questionMarkPos == len(s)-1 {
		return strings.ToLower(strings.ReplaceAll(s[0:len(s)-1], " ", ""))
	}
	return ""
}
