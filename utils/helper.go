package utils

import "strings"

func Capitalise(str string) string {
	firstLetter := strings.ToUpper(str[:1])
	others := str[1:]

	return firstLetter + others
}
