package parse_js

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

var EX_EOF = "END OF FILE"

//func  tokenizer(TEXT string)  {

type Tokenizer struct {
	text            string   // : TEXT.replace(/\r\n?|[\n\u2028\u2029]/g, "\n").replace(/^\uFEFF/, ''),
	pos             int64    // : 0,
	tokpos          int64    // : 0,
	line            int64    // : 0,
	tokline         int64    // : 0,
	col             int64    // : 0,
	tokcol          int64    // : 0,
	newline_before  bool     // : false,
	regex_allowed   bool     // : false,
	comments_before []string //
}

func NewTokenizer(str string) *Tokenizer {
	t := new(Tokenizer)
	reg, err := regexp.Compile(`/\r\n?|[\n\u2028\u2029]/g`)
	if err != nil {
		panic(err)
	}
	str = reg.ReplaceAllString(str, "\n")

	reg, err = regexp.Compile(`/^\uFEFF/`)
	if err != nil {
		panic(err)
	}

	str = reg.ReplaceAllString(str, "")

	t.text = str
	return t
}

func (t *Tokenizer) peek() string {
	return t.text[t.pos : t.pos+1]
}

func (t *Tokenizer) next(signal_eof bool, in_string bool) string {
	t.pos++
	var ch = t.text[t.pos : t.pos+1]
	if signal_eof && ch == "" {
		panic(EX_EOF)
	}
	if ch == "\n" {
		t.newline_before = t.newline_before || in_string
		t.line++
		t.col = 0
	} else {
		t.col++
	}
	return ch
}

func (t *Tokenizer) eof() bool {
	return t.peek() == ""

}

func (t *Tokenizer) find(what string) int64 {
	pos := strings.IndexRune(t.text[t.pos:], []rune(what)[0])
	//	var pos = t.text.indexOf(what, t.pos)

	return int64(pos)
}

func (t *Tokenizer) start_token() {
	t.tokline = t.line
	t.tokcol = t.col
	t.tokpos = t.pos
}

type token struct {
	typ             string //: tp,
	value           string //: value,
	line            int64  //: t.tokline,
	col             int64  //: t.tokcol,
	pos             int64  //: t.tokpos,
	endpos          int64  //: t.pos,
	nlb             bool   // : t.newline_before
	comments_before []string
}

func (t *Tokenizer) token(tp string, value string, is_comment bool) *token {
	t.regex_allowed = (tp == "operator" && !HOP(UNARY_POSTFIX, value)) ||
		(tp == "keyword" && HOP(KEYWORDS_BEFORE_EXPRESSION, value)) ||
		(tp == "punc" && HOP(PUNC_BEFORE_EXPRESSION, value))
	ret := token{
		typ:    tp,
		value:  value,
		line:   t.tokline,
		col:    t.tokcol,
		pos:    t.tokpos,
		endpos: t.pos,
		nlb:    t.newline_before,
	}
	if (!is_comment) {
		ret.comments_before = t.comments_before
		t.comments_before = []string{}
		// make note of any newlines in the comments that came before
		//for (var i = 0, len = ret.comments_before.length i < len i++) {
		//ret.nlb = ret.nlb || ret.comments_before[i].nlb
		//}
		//}
		//for _,v := range ret.comments_before {
		//	ret.nlb = ret.nlb // || v.nlb
		//}
		t.newline_before = false
	}
	return &ret
}

func (t *Tokenizer) skip_whitespace() {
	for HOP(WHITESPACE_CHARS, t.peek()) {
		str := t.next(false, false)
		log.Println(str)
	}
}

func (t *Tokenizer) read_while(pred func(ch string, i int) bool) string {
	ret := ""
	ch := t.peek()
	i := 0
	for ch != "" && pred(ch, i) {
		i++
		ret += t.next(false, false)
		ch = t.peek()
	}
	return ret

}

func (t *Tokenizer) parse_error(err string) {
	log.Panic(err, t)
}

func (t *Tokenizer) read_num(prefix string) *token {
	has_e := false
	after_e := false
	has_x := false
	has_dot := prefix == "."

	var num = t.read_while(func(ch string, i int) bool {
		if ch == "x" || ch == "X" {
			if has_x {
				return false
			}
			has_x = true

			return has_x
		}
		if (!has_x && (ch == "E" || ch == "e")) {
			if (has_e) {
				return false
			}
			has_e = true
			after_e = true
			return true
		}
		if ch == "-" {
			if after_e || (i == 0 && prefix != "") {
				return true
			}
			return false
		}
		if ch == "+" {
			return after_e
		}
		after_e = false
		if ch == "." {
			if !has_dot && !has_x && !has_e {
				has_dot = true
				return true
			}
			return false
		}
		return IsAlphanumericChar(ch)
	})

	if prefix != "" {
		num = prefix + num
	}
	var valid = ParseJsNumber(num)
	if valid != nil {
		return t.token("num", valid, false)
	} else {
		panic("Invalid syntax: " + num)
	}
	return nil
}

func (t *Tokenizer) read_escaped_char(in_string bool) string {
	var ch = t.next(true, in_string)
	switch ch {
	case "n":
		return "\n"
	case "r":
		return "\r"
	case "t":
		return "\t"
	case "b":
		return "\b"
	case "v":
		return "\u000b"
	case "f":
		return "\f"
	case "0":
		return "\0"
	case "x":
		return strings.fromCharCode(t.hex_bytes(2))
	case "u":
		return strings.fromCharCode(t.hex_bytes(4))
	case "\n":
		return ""
	default:
		return ch
	}
}

func (t *Tokenizer) hex_bytes(n int) int64 {
	var num int64 = 0
	for
	(n > 0 --
	n) {
		var digit = parseInt(t.next(true, false), 16)
		if (isNaN(digit))
			parse_error("Invalid hex-character pattern in string")
		num = (num << 4) | digit
	}
	return num
}

func (t *Tokenizer) read_string() *token {
	return t.with_eof_error("Unterminated string constant", func() *token {
		quote := t.next(false, false)
		ret := ""
		for true {
			var ch = t.next(true, false)
			if (ch == "\\") {
				// read OctalEscapeSequence (XXX: deprecated if "strict mode")
				// https://github.com/mishoo/UglifyJS/issues/178
				octal_len := 0
				first := ""
				ch = t.read_while(func(ch string, i int) bool {
					if (ch >= "0" && ch <= "7") {
						if (first == "") {
							first = ch
							octal_len++
							return true
						}
						if (first <= "3" && octal_len <= 2) {
							octal_len++
							return true
						}
						if (first >= "4" && octal_len <= 1) {
							octal_len++
							return true
						}
					}
					return false
				})
				if (octal_len > 0) {
					ch = String.fromCharCode(parseInt(ch, 8))
				} else {
					ch = t.read_escaped_char(true)
				}
			} else if (ch == quote) {
				break
			} else if (ch == "\n") {
				log.Panic(EX_EOF)
			}

			ret += ch
		}
		return t.token("string", ret, false)
	})
}

func (t *Tokenizer) read_line_comment() *token {
	t.next(false, false)
	i := t.find("\n")
	ret := ""
	if (i == -1) {
		ret = t.text.substr(t.pos)
		t.pos = t.text.length
	} else {
		ret = t.text.substring(S.pos, i)
		t.pos = i
	}
	return t.token("comment1", ret, true)
}

func (t *Tokenizer) read_multiline_comment() *token {
	t.next(false, false)
	return t.with_eof_error("Unterminated multiline comment", func() *token {
		i := t.find("*/")
		text := t.text[t.pos:i]
		t.pos = i + 2
		t.line += text.split("\n").length - 1
		t.newline_before = t.newline_before || text.indexOf("\n") >= 0

		// https://github.com/mishoo/UglifyJS/issues/#issue/100
		if ( / ^@cc_on / i.test(text)) {
			warn("WARNING: at line " + S.line)
			warn("*** Found \"conditional comment\": " + text)
			warn("*** UglifyJS DISCARDS ALL COMMENTS.  This means your code might no longer work properly in Internet Explorer.")
		}

		return t.token("comment2", text, true)
	})
}

func (t *Tokenizer) read_name() string {
	backslash := false
	name := ""
	ch := ""
	escaped := false
	hex := ""
	for true {
		ch = t.peek()
		if ch == "" {
			break
		}

		if (!backslash) {
			if (ch == "\\") {
				escaped = true
				backslash = true
				t.next(false, false)
			} else {
				if (IsIdentifierChar(ch)) {
					name += t.next(false, false)
				} else {
					break
				}
			}
		} else {
			if (ch != "u") {
				log.Panic("Expecting UnicodeEscapeSequence -- uXXXX")
			}
			ch = t.read_escaped_char(false)
			if (!IsIdentifierChar(ch)) {
				log.Panic("Unicode char: ", ch, " is not valid in identifier")
			}
			name += ch
			backslash = false
		}
	}
	if (HOP(KEYWORDS, name) && escaped) {
		hex = name.charCodeAt(0).toString(16).toUpperCase()
		name = "\\u" + "0000".substr(hex.length) + hex + name.slice(1)
	}
	return name
}

func (t *Tokenizer) read_regexp(regexp *string) *token {
	return t.with_eof_error("Unterminated regular expression", func() *token {
		prev_backslash := false
		ch := ""
		in_class := false

		for true {
			ch = t.next(true, false)
			if ch == "" {
				break
			}
			if prev_backslash {
				*regexp += "\\" + ch
				prev_backslash = false
			} else if (ch == "[") {
				in_class = true
				*regexp += ch
			} else if (ch == "]" && in_class) {
				in_class = false
				*regexp += ch
			} else if (ch == "/" && !in_class) {
				break
			} else if (ch == "\\") {
				prev_backslash = true
			} else {
				*regexp += ch
			}
		}
		var mods = t.read_name()
		return t.token("regexp", []string{*regexp, mods}, false)
	})
}

func (t *Tokenizer) read_operator(prefix string) *token {
	if prefix != "" {
		t.token("operator", t.grow(prefix), false)
	}
	return t.token("operator", t.grow(t.next(false, false)), false)
}
func (t *Tokenizer) grow(op string) string {
	if t.peek() != "" {
		return op
	}
	var bigger = op + t.peek()
	if (HOP(OPERATORS, bigger)) {
		t.next(false, false)
		return t.grow(bigger)
	} else {
		return op
	}
}
func (t *Tokenizer) handle_slash() *token {
	t.next(false, false)
	var regex_allowed = t.regex_allowed
	switch (t.peek()) {
	case "/":
		t.comments_before.push(t.read_line_comment())
		t.regex_allowed = regex_allowed
		return t.next_token(nil)
	case "*":
		t.comments_before.push(t.read_multiline_comment())
		t.regex_allowed = regex_allowed
		return t.next_token(nil)
	}
	if t.regex_allowed {
		return t.read_regexp(nil)
	}
	return t.read_operator("/")
}

func (t *Tokenizer) handle_dot() *token {
	t.next(false, false)
	if IsDigit(t.peek()) {

		return t.read_num(".")
	}

	return t.token("punc", ".", false)

}

func (t *Tokenizer) read_word() *token {
	var word = t.read_name()
	if !HOP(KEYWORDS, word) {
		return t.token("name", word, false)
	}
	if HOP(OPERATORS, word) {
		return t.token("operator", word, false)
	}
	if HOP(KEYWORDS_ATOM, word) {
		return t.token("atom", word, false)
	}
	return t.token("keyword", word, false)
}

func (t *Tokenizer) with_eof_error(eof_error string, cont func() *token) *token {
	return cont()
}

func (t *Tokenizer) next_token(force_regexp *string) *token {
	if (force_regexp != nil) {
		return t.read_regexp(force_regexp)
	}
	t.skip_whitespace()
	t.start_token()
	var ch = t.peek()
	if (ch == "") {
		return t.token("eof", "", false)
	}
	if (IsDigit(ch)) {
		return t.read_num("")
	}
	if ch == `"` || ch == "'" {
		return t.read_string()
	}
	if (HOP(PUNC_CHARS, ch)) {
		return t.token("punc", t.next(false,false),false)
	}
	if (ch == ".") {
		return t.handle_dot()
	}
	if (ch == "/") {
		return t.handle_slash()
	}
	if (HOP(OPERATOR_CHARS, ch)) {
		return t.read_operator("")
	}
	if (ch == "\\" || IsIdentifierStart(ch)){
	return 	t.read_word()
	}
	log.Panic("Unexpected character '" + ch + "'")
	return nil
}
