package parse_js

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

var EX_EOF = "END OF FILE"

type Tokenizer struct {
	Text           string
	Pos            int
	Tokpos         int
	Line           int
	Tokline        int
	Col            int
	Tokcol         int
	NewlineBefore  bool
	RegexAllowed   bool
	commentsBefore []*Token
}

func NewTokenizer(str string) *Tokenizer {
	t := new(Tokenizer)
	reg, err := regexp.Compile("/\r\n?|[\n\u2028\u2029]/g")
	if err != nil {
		panic(err)
	}
	str = reg.ReplaceAllString(str, "\n")

	reg, err = regexp.Compile("/^\uFEFF/")
	if err != nil {
		panic(err)
	}

	str = reg.ReplaceAllString(str, "")

	t.Text = str
	return t
}

func (t *Tokenizer) Peek() string {
	if len(t.Text) == t.Pos {
		return ""
	}
	return t.Text[t.Pos : t.Pos+1]
}

func (t *Tokenizer) Next(signal_eof bool, in_string bool) string {
	var ch = t.Peek()
	t.Pos++
	if signal_eof && ch == "" {
		panic(EX_EOF)
	}
	if ch == "\n" {
		t.NewlineBefore = t.NewlineBefore || in_string
		t.Line++
		t.Col = 0
	} else {
		t.Col++
	}
	return ch
}

func (t *Tokenizer) Eof() bool {
	return t.Peek() == ""
}

func (t *Tokenizer) Find(what string) int {
	pos := strings.IndexRune(t.Text[t.Pos:], []rune(what)[0])
	return pos
}

func (t *Tokenizer) startToken() {
	t.Tokline = t.Line
	t.Tokcol = t.Col
	t.Tokpos = t.Pos
}

type Token struct {
	Typ            string
	Value          interface{}
	Line           int
	Col            int
	Pos            int
	Endpos         int
	Nlb            bool
	CommentsBefore []*Token
}

func (t *Tokenizer) Token(tp string, value interface{}, is_comment bool) *Token {
	t.RegexAllowed = (tp == "operator" && !HOP(UNARY_POSTFIX, value.(string))) ||
		(tp == "keyword" && HOP(KEYWORDS_BEFORE_EXPRESSION, value.(string))) ||
		(tp == "punc" && HOP(PUNC_BEFORE_EXPRESSION, value.(string)))
	ret := Token{
		Typ:    tp,
		Value:  value,
		Line:   t.Tokline,
		Col:    t.Tokcol,
		Pos:    t.Tokpos,
		Endpos: t.Pos,
		Nlb:    t.NewlineBefore,
	}
	if !is_comment {
		ret.CommentsBefore = t.commentsBefore
		t.commentsBefore = []*Token{}
		// make note of any newlines in the comments that came before
		//for (var i = 0, len = ret.comments_before.length i < len i++) {
		//ret.nlb = ret.nlb || ret.comments_before[i].nlb
		//}
		//}
		//for _,v := range ret.comments_before {
		//	ret.nlb = ret.nlb // || v.nlb
		//}
		t.NewlineBefore = false
	}
	return &ret
}

func (t *Tokenizer) SkipWhitespace() {
	for HOP(WHITESPACE_CHARS, t.Peek()) {
		t.Next(false, false)
	}
}

func (t *Tokenizer) ReadWhile(pred func(ch string, i int) bool) string {
	ret := ""
	ch := t.Peek()
	i := 0
	for ch != "" && pred(ch, i) {
		i++
		ret += t.Next(false, false)
		ch = t.Peek()
	}
	return ret

}

func (t *Tokenizer) ReadNum(prefix string) *Token {
	has_e := false
	after_e := false
	has_x := false
	has_dot := prefix == "."

	var num = t.ReadWhile(func(ch string, i int) bool {
		if ch == "x" || ch == "X" {
			if has_x {
				return false
			}
			has_x = true

			return has_x
		}
		if !has_x && (ch == "E" || ch == "e") {
			if has_e {
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
		return t.Token("num", valid, false)
	} else {
		panic("Invalid syntax: " + num)
	}
	return nil
}

func (t *Tokenizer) ReadEscapedChar(in_string bool) string {
	var ch = t.Next(true, in_string)
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
		return `\0`
	case "x":
		return string(t.hexBytes(2))
	case "u":
		return string(t.hexBytes(4))
	case "\n":
		return ""
	default:
		return ch
	}
}

func (t *Tokenizer) hexBytes(n int) int64 {
	var num int64 = 0
	for ; n > 0; n-- {

		digit, err := strconv.ParseInt(t.Next(true, false), 2, 16)

		if err != nil {
			log.Panic("Invalid hex-character pattern in string")
		}
		num = (num << 4) | digit
	}
	return num
}

func (t *Tokenizer) ReadString() *Token {
	return t.WithEofError("Unterminated string constant", func() *Token {
		quote := t.Next(false, false)
		ret := ""
		for true {
			var ch = t.Next(true, false)
			if ch == "\\" {
				// read OctalEscapeSequence (XXX: deprecated if "strict mode")
				// https://github.com/mishoo/UglifyJS/issues/178
				octal_len := 0
				first := ""
				ch = t.ReadWhile(func(ch string, i int) bool {
					if ch >= "0" && ch <= "7" {
						if first == "" {
							first = ch
							octal_len++
							return true
						}
						if first <= "3" && octal_len <= 2 {
							octal_len++
							return true
						}
						if first >= "4" && octal_len <= 1 {
							octal_len++
							return true
						}
					}
					return false
				})
				if octal_len > 0 {

					chNum, err := strconv.ParseInt(ch, 2, 8)
					if err != nil {
						log.Println("parse_js_number", ch)
					}

					ch = strconv.FormatInt(chNum, 10)
					//	ch = string(parseInt(ch, 8))
				} else {
					ch = t.ReadEscapedChar(true)
				}
			} else if ch == quote {
				break
			} else if ch == "\n" {
				log.Panic(EX_EOF)
			}

			ret += ch
		}
		return t.Token("string", ret, false)
	})
}

func (t *Tokenizer) ReadLineComment() *Token {
	t.Next(false, false)
	i := t.Find("\n")
	ret := ""
	if i == -1 {
		ret = t.Text[t.Pos:]
		t.Pos = len(t.Text)
	} else {
		ret = t.Text[t.Pos:i]
		t.Pos = int(i)
	}
	return t.Token("comment1", ret, true)
}

func (t *Tokenizer) ReadMultilineComment() *Token {
	t.Next(false, false)
	return t.WithEofError("Unterminated multiline comment", func() *Token {
		i := t.Find("*/")
		text := t.Text[t.Pos:i]
		t.Pos = 2 + i
		t.Line += len(strings.Split(text, "\n")) - 1
		t.NewlineBefore = t.NewlineBefore || strings.IndexRune(text, '\n') >= 0

		// https://github.com/mishoo/UglifyJS/issues/#issue/100
		//if ( / ^@cc_on / i.test(text)) {
		//	warn("WARNING: at line " + S.line)
		//	warn("*** Found \"conditional comment\": " + text)
		//	warn("*** UglifyJS DISCARDS ALL COMMENTS.  This means your code might no longer work properly in Internet Explorer.")
		//}

		return t.Token("comment2", text, true)
	})
}

func (t *Tokenizer) ReadName() string {
	backslash := false
	name := ""
	ch := ""
	escaped := false
	for true {
		ch = t.Peek()

		if ch == "" {
			break
		}

		if !backslash {
			if ch == "\\" {
				escaped = true
				backslash = true
				ch = t.Next(false, false)
			} else {
				if IsIdentifierChar(ch) {
					name += t.Next(false, false)
				} else {
					break
				}
			}
		} else {
			log.Println(ch)
			if ch != "u" {
				log.Panic("Expecting UnicodeEscapeSequence -- uXXXX")
			}
			ch = t.ReadEscapedChar(false)
			if !IsIdentifierChar(ch) {
				log.Panic("Unicode char: ", ch, " is not valid in identifier")
			}
			name += ch
			backslash = false
		}
	}

	if HOP(KEYWORDS, name) && escaped {
		log.Println("HOP(KEYWORDS")
		//hex = name.charCodeAt(0).toString(16).toUpperCase()
		//name = "\\u" + "0000".substr(hex.length) + hex + name.slice(1)

		log.Fatalln(name)
	}
	return name
}

func (t *Tokenizer) ReadRegexp(regexp *string) *Token {
	return t.WithEofError("Unterminated regular expression", func() *Token {
		prev_backslash := false
		ch := ""
		in_class := false

		for true {
			ch = t.Next(true, false)
			if ch == "" {
				break
			}
			if prev_backslash {
				*regexp += "\\" + ch
				prev_backslash = false
			} else if ch == "[" {
				in_class = true
				*regexp += ch
			} else if ch == "]" && in_class {
				in_class = false
				*regexp += ch
			} else if ch == "/" && !in_class {
				break
			} else if ch == "\\" {
				prev_backslash = true
			} else {
				*regexp += ch
			}
		}
		var mods = t.ReadName()
		return t.Token("regexp", []string{*regexp, mods}, false)
	})
}

func (t *Tokenizer) ReadOperator(prefix string) *Token {
	if prefix != "" {
		t.Token("operator", t.Grow(prefix), false)
	}
	return t.Token("operator", t.Grow(t.Next(false, false)), false)
}
func (t *Tokenizer) Grow(op string) string {
	if t.Peek() != "" {
		return op
	}
	var bigger = op + t.Peek()
	if HOP(OPERATORS, bigger) {
		t.Next(false, false)
		return t.Grow(bigger)
	} else {
		return op
	}
}
func (t *Tokenizer) HandleSlash() *Token {
	t.Next(false, false)
	var regexAllowed = t.RegexAllowed
	switch t.Peek() {
	case "/":
		t.commentsBefore = append(t.commentsBefore, t.ReadLineComment())
		t.RegexAllowed = regexAllowed
		return t.NextToken(nil)
	case "*":
		t.commentsBefore = append(t.commentsBefore, t.ReadLineComment())
		t.RegexAllowed = regexAllowed
		return t.NextToken(nil)
	}
	if t.RegexAllowed {
		return t.ReadRegexp(nil)
	}
	return t.ReadOperator("/")
}

func (t *Tokenizer) HandleDot() *Token {
	t.Next(false, false)
	if IsDigit(t.Peek()) {
		return t.ReadNum(".")
	}

	return t.Token("punc", ".", false)
}

func (t *Tokenizer) ReadWord() *Token {
	var word = t.ReadName()

	if !HOP(KEYWORDS, word) {
		return t.Token("name", word, false)
	}

	if HOP(OPERATORS, word) {
		return t.Token("operator", word, false)
	}

	if HOP(KEYWORDS_ATOM, word) {
		return t.Token("atom", word, false)
	}

	return t.Token("keyword", word, false)
}

func (t *Tokenizer) WithEofError(eof_error string, cont func() *Token) *Token {
	tkn := cont()
	if tkn == nil {
		log.Fatalln(eof_error, t)
	}
	return tkn
}

func (t *Tokenizer) NextToken(force_regexp *string) *Token {
	if force_regexp != nil {
		return t.ReadRegexp(force_regexp)
	}
	t.SkipWhitespace()
	t.startToken()
	ch := t.Peek()

	if ch == "" {
		return t.Token("eof", "", false)
	}
	if IsDigit(ch) {
		return t.ReadNum("")
	}
	if ch == `"` || ch == "'" {
		return t.ReadString()
	}
	if HOP(PUNC_CHARS, ch) {
		return t.Token("punc", t.Next(false, false), false)
	}
	if ch == "." {
		return t.HandleDot()
	}
	if ch == "/" {
		return t.HandleSlash()
	}
	if HOP(OPERATOR_CHARS, ch) {
		return t.ReadOperator("")
	}
	if ch == "\\" || IsIdentifierStart(ch) {
		return t.ReadWord()
	}
	log.Panic("Unexpected character '" + ch + "'")
	return nil
}
