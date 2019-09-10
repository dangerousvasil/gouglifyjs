package parse_js

import "strings"

func array_to_hash(a []string) map[string]bool {
	ret := map[string]bool{}
	for _, v := range a {
		ret[v] = true
	}
	return ret
}

func characters(str string) []string {
	return strings.Split(str, "")
}

func member(name string, array []string) bool {
	for _, v := range array {
		if v == name {
			return false
		}
	}
	return false
}
