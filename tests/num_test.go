package _go

import (
	parse_js "gouglifyjs/src/parse-js"
	"log"
	"testing"
)

func TestNum(t *testing.T) {
	str := "/a/ / /b/;"

	log.Println(str)
	tok := parse_js.NewTokenizer(str)
	for true {
		txt := tok.NextToken(nil)
		if txt == nil {
			return
		}
		log.Println(txt)
	}
}
