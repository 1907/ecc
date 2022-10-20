package importing

import (
	"hash/crc32"
	"regexp"
	"strconv"
	"strings"
)

func Include(arr []string, str string) bool {
	if len(arr) == 0 {
		return false
	}
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func Prefix(name string) string {
	r := regexp.MustCompile("^([0-9a-zA-Z]+)-")
	params := r.FindStringSubmatch(name)
	if len(params) > 1 && params[1] != "" {
		return params[1]
	}
	return "unknown"
}

func Hash(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return 0
}

func HashString(str ...string) string {
	_str := strings.Join(str, "@@")
	return strconv.Itoa(Hash(_str))
}

func ToStringSlice(items []interface{}) []string {
	var str []string
	for _, item := range items {
		str = append(str, item.(string))
	}
	return str
}
