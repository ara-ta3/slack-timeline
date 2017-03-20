package slack

import "regexp"

var format = regexp.MustCompile("<@([a-zA-Z0-9]+)\\|{1}?(.+?)>+?")

func ReplaceIdFormatToName(s string) string {
	return format.ReplaceAllString(s, "$2")
}
