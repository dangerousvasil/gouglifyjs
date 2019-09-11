package main

import (
	parse_js "gouglifyjs/src/parse-js"
	"io/ioutil"
	"log"
)

func main() {
	//log.Println(parse_js.UNARY_PREFIX)
	//log.Println(parse_js.UNARY_POSTFIX)
	//log.Println(parse_js.STATEMENTS_WITH_LABELS)
	//log.Println(parse_js.ATOMIC_START_TOKEN)
	//log.Println(parse_js.ASSIGNMENT)
	//log.Println(parse_js.PRECEDENCE)
	//log.Println(parse_js.KEYWORDS)
	//log.Println(parse_js.RESERVED_WORDS)
	//log.Println(parse_js.KEYWORDS_BEFORE_EXPRESSION)
	//log.Println(parse_js.KEYWORDS_ATOM)
	//log.Println(parse_js.OPERATOR_CHARS)
	//log.Println(parse_js.RE_HEX_NUMBER)
	//log.Println(parse_js.RE_OCT_NUMBER)
	//log.Println(parse_js.RE_DEC_NUMBER)
	//log.Println(parse_js.OPERATORS)
	//log.Println(parse_js.WHITESPACE_CHARS)
	//log.Println(parse_js.PUNC_BEFORE_EXPRESSION)
	//log.Println(parse_js.PUNC_CHARS)
	//log.Println(parse_js.REGEXP_MODIFIERS)
	//log.Println(parse_js.UNICODE)
	//log.Println(parse_js.UNICODE)
	jsByt, err := ioutil.ReadFile(`/home/pvv/IdeaProjects/gouglifyjs/test/unit/scripts.js`)
	if err != nil {
		log.Fatalln(err)
	}
	t := parse_js.NewTokenizer(string(jsByt))
	for true {
		txt := t.NextToken(nil)
		if txt == nil {
			break
		}
		log.Println("-=-=-=-=-=-=-=-")
		log.Println(txt)

	}
}
