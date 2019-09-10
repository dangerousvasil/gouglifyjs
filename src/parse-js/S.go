package parse_js

//
//type S struct {
//	Input        *string  // : typeof $TEXT == "string" ? tokenizer($TEXT, true) : $TEXT,
//	Token        *string  // : null,
//	Prev         string   // : null,
//	Peeked       *string  // : null,
//	InFunction   int      // : 0,
//	InDirectives bool     // : true,
//	InLoop       int      // : 0,
//	Labels       []string // : []
//}
//
//func (S *S) next() {
//	S.Prev = S.Token;
//	if S.Peeked != nil {
//		S.Token = S.Peeked
//		S.Peeked = nil
//	} else {
//		S.Token = S.Input();
//	}
//	S.InDirectives = S.InDirectives && (
//		S.Token.
//	type == "string" || is("punc", ";")
//	)
//	return S.Token
//}
//
//func (S *S) peek() bool {
//	return S.Peeked || (S.Peeked = S.Input())
//}
