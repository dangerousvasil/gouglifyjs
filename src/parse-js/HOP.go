package parse_js

func HOP(obj map[string]bool, value string) bool {
	_, ok := obj[value]
	return ok
}
