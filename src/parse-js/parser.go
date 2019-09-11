package parse_js

import (
	"fmt"
	"go/parser"
	"log"
)

//func  NodeWithToken(str, start, end int) {
//this.name = str;
//this.start = start;
//this.end = end;
//};

//var S = {
//input         : typeof $TEXT == "string" ? tokenizer($TEXT, true) : $TEXT,
//token         : null,
//prev          : null,
//peeked        : null,
//in_function   : 0,
//in_directives : true,
//in_loop       : 0,
//labels        : []
//};

type Parser struct {
	input         Tokenizer // : typeof $TEXT == "string" ? tokenizer($TEXT, true) : $TEXT,
	token         *Token    // : null,
	Prev          *Token    //: null,
	peeked        *Token    //: null,
	in_function   int       //: 0,
	in_directives bool      //: true,
	In_loop       int       //: 0,
	labels        []string  //: []

	exigent_mode bool
	embed_tokens bool
}

func NewParser(text string, exigent_mode, embed_tokens bool) *Parser {
	p := new(Parser)
	p.exigent_mode = exigent_mode
	p.embed_tokens = embed_tokens
	return p
}

//S.token = next();

func (p *Parser) is(typ, value string) bool {
	return IsToken(p.token, typ, value)
}

func (p *Parser) peek() *Token {
	if p.peeked == nil {
		p.peeked = p.input.NextToken(nil)
	}
	return p.peeked
}

func (p *Parser) next() *Token {
	p.Prev = p.token
	if p.peeked == nil {
		p.token = p.peeked
		p.peeked = nil
	} else {
		p.token = p.input.NextToken(nil)
	}
	p.in_directives = p.in_directives && (
		p.token.Typ == "string" || p.is("punc", ";"));
	return p.token;
};

func (p *Parser) prev() *Token {
	return p.Prev
}

func (p *Parser) croak(msg string, line, col, pos int) {
	log.Panic(msg, p)
	//var ctx = p.input.context()
	//js_error(msg,
	//	line != null ? line:
	//ctx.tokline,
	//	col != null ? col:
	//ctx.tokcol,
	//	pos != null ? pos:
	//ctx.tokpos);
}

func (p *Parser) token_error(token *Token, msg string) {
	p.croak(msg, token.Line, token.Col, 0)
}

func (p *Parser) unexpected(token *Token) {
	if token == nil {
		token = p.token;
	}
	p.token_error(token, fmt.Sprintf("Unexpected token: %s  (%v)", token.Typ, token.Value))
}

func (p *Parser) expect_token(typ, val string) *Token {
	if p.is(typ, val) {
		return p.next()
	}
	p.token_error(p.token, fmt.Sprintf("Unexpected token %s expected %s", p.token.Typ, typ))
	return nil
}

func (p *Parser) expect(punc string) *Token {
	return p.expect_token("punc", punc)
}

func (p *Parser) can_insert_semicolon() bool{
	return !p.exigent_mode && (
		p.token.Nlb || p.is("eof","") || p.is("punc", "}")
	)
}

func (p *Parser) semicolon() {
	if p.is("punc", ";") {
		p.next();
	} else {
		if !p.can_insert_semicolon() {
			p.unexpected(nil)
		}
	};
}

//func (p *Parser) as() {
//	return slice(arguments);
//};

func (p *Parser) parenthesised() {
	p.expect("(");
	var ex = expression();
	p.expect(")");
	return ex;
};
//
//func (p *Parser) add_tokens(str, start, end) {
//	return str
//	instanceof
//	NodeWithToken ? str:
//	new
//	NodeWithToken(str, start, end);
//};

func (p *Parser) maybe_embed_tokens(parser) {
	if (p.embed_tokens)
	return func ()
	{
		var start = p.token;
		var ast = parser.apply(this, arguments);
		ast[0] = add_tokens(ast[0], start, prev());
		return ast;
	};
	else return p;
};

var statement = maybe_embed_tokens(func () {
if (is("operator", "/") || is("operator", "/=")) {
S.peeked = null;
S.token = S.input(S.token.value.substr(1)); // force regexp
}
switch (S.token.typ
) {
case "string":
var dir = S.in_directives, stat = simple_statement();
if (dir && stat[1][0] == "string" && !is("punc", ","))
return as("directive", stat[1][1]);
return stat;
case "num":
case "regexp":
case "operator":
case "atom":
return simple_statement();

case "name":
return is_token(peek(), "punc", ":")
? labeled_statement(prog1(S.token.value, next, next)): simple_statement();

case "punc":
switch (S.token.value) {
case "{":
return as("block", block_());
case "[":
case "(":
return simple_statement();
case ";":
next();
return as("block");
default:
unexpected();
}

case "keyword":
switch (prog1(S.token.value, next)) {
case "break":
return break_cont("break");

case "continue":
return break_cont("continue");

case "debugger":
semicolon();
return as("debugger");

case "do":
return (function(body){
expect_token("keyword", "while");
return as("do", prog1(parenthesised, semicolon), body);
})(in_loop(statement));

case "for":
return for_();

case "function":
return function_(true);

case "if":
return if_();

case "return":
if (S.in_function == 0)
croak("'return' outside of function");
return as("return",
is("punc", ";")
? (next(), null)
: can_insert_semicolon()
? null: prog1(expression, semicolon));

case "switch":
return as("switch", parenthesised(), switch_block_());

case "throw":
if (S.token.nlb)
croak("Illegal newline after 'throw'");
return as("throw", prog1(expression, semicolon));

case "try":
return try_();

case "var":
return prog1(var_, semicolon);

case "const":
return prog1(const_, semicolon);

case "while":
return as("while", parenthesised(), in_loop(statement));

case "with":
return as("with", parenthesised(), statement());

default:
unexpected();
}
}
});

func (p *Parser) labeled_statement(label) {
	S.labels.push(label);
	var start = S.token, stat = statement();
	if (exigent_mode && !HOP(STATEMENTS_WITH_LABELS, stat[0]))
		unexpected(start);
	S.labels.pop();
	return as("label", label, stat);
};

func (p *Parser) simple_statement() {
	return as("stat", prog1(expression, semicolon));
};

func (p *Parser) break_cont(

type
) {
var name;
if (!can_insert_semicolon()) {
name = is("name") ? S.token.value: null;
}
if (name != null) {
next();
if (!member(name, S.labels))
croak("Label " + name + " without matching loop or statement");
}
else if (S.in_loop == 0)
croak(type + " not inside a loop or switch");
semicolon();
return as(type, name);
};

func (p *Parser) for_() {
	expect("(");
	var init = null;
	if (!is("punc", ";")) {
		init = is("keyword", "var")
		? (next(), var_(true)): expression(true, true);
		if (is("operator", "in")) {
			if (init[0] == "var" && init[1].length > 1)
				croak("Only one variable declaration allowed in for..in loop");
			return for_in(init);
		}
	}
	return regular_for(init);
};

func (p *Parser) regular_for(init) {
	expect(";");
	var test = is("punc", ";") ? null:
	expression();
	expect(";");
	var step = is("punc", ")") ? null:
	expression();
	expect(")");
	return as("for", init, test, step, in_loop(statement));
};

func (p *Parser) for_in(init) {
	var lhs = init[0] == "var" ? as("name", init[1][0]) : init;
	next();
	var obj = expression();
	expect(")");
	return as("for-in", init, lhs, obj, in_loop(statement));
};

var function_ = function(in_statement) {
var name = is("name") ? prog1(S.token.value, next): null;
if (in_statement && !name)
unexpected();
expect("(");
return as(in_statement ? "defun": "function",
name,
// arguments
(function(first, a){
while (!is("punc", ")")) {
if (first) first = false; else expect(",");
if (!is("name")) unexpected();
a.push(S.token.value);
next();
}
next();
return a;
})(true, []),
// body
(function(){
++S.in_function;
var loop = S.in_loop;
S.in_directives = true;
S.in_loop = 0;
var a = block_();
--S.in_function;
S.in_loop = loop;
return a;
})());
};

func (p *Parser) if_() {
	var cond = parenthesised(), body = statement(), belse;
	if (is("keyword", "else")) {
		next();
		belse = statement();
	}
	return as("if", cond, body, belse);
};

func (p *Parser) block_() {
	expect("{");
	var a = [];
	while(!is("punc", "}"))
	{
		if (is("eof")) unexpected();
		a.push(statement());
	}
	next();
	return a;
};

var switch_block_ = curry(in_loop, function(){
expect("{");

var a = [], cur = null;
while (!is("punc", "}")) {
if (is("eof")) unexpected();
if (is("keyword", "case")) {
next();
cur = [];
a.push([ expression(), cur ]);
expect(":");
}
else if (is("keyword", "default")) {
next();
expect(":");
cur = [];
a.push([ null, cur ]);
}
else {
if (!cur) unexpected();
cur.push(statement());
}
}
next();
return a;
});

func (p *Parser) try_() {
	var body = block_(), bcatch, bfinally;
	if (is("keyword", "catch")) {
		next();
		expect("(");
		if (!is("name"))
			croak("Name expected");
		var name = S.token.value;
		next();
		expect(")");
		bcatch = [ name, block_() ];
}
if (is("keyword", "finally")) {
next();
bfinally = block_();
}
if (!bcatch && !bfinally)
croak("Missing catch/finally blocks");
return as("try", body, bcatch, bfinally);
};

func (p *Parser) vardefs(no_in) {
	var a = [];
	for (; ;) {
		if (!is("name"))
			unexpected();
		var name = S.token.value;
		next();
		if (is("operator", "=")) {
			next();
			a.push([ name, expression(false, no_in) ]);
} else {
a.push([ name ]);
}
if (!is("punc", ","))
break;
next();
}
return a;
};

func (p *Parser) var_(no_in) {
	return as("var", vardefs(no_in));
};

func (p *Parser) const_() {
	return as("const", vardefs());
};

func (p *Parser) new_() {
	var newexp = expr_atom(false), args;
	if (is("punc", "(")) {
		next();
		args = expr_list(")");
	} else {
		args = [];
	}
	return subscripts(as("new", newexp, args), true);
};

var expr_atom = maybe_embed_tokens(function(allow_calls) {
if (is("operator", "new")) {
next();
return new_();
}
if (is("punc")) {
switch (S.token.value) {
case "(":
next();
return subscripts(prog1(expression, curry(expect, ")")), allow_calls);
case "[":
next();
return subscripts(array_(), allow_calls);
case "{":
next();
return subscripts(object_(), allow_calls);
}
unexpected();
}
if (is("keyword", "function")) {
next();
return subscripts(function_(false), allow_calls);
}
if (HOP(ATOMIC_START_TOKEN, S.token.type
)) {
var atom = S.token.type == "regexp"
? as("regexp", S.token.value[0], S.token.value[1]): as(S.token.type, S.token.value);
return subscripts(prog1(atom, next), allow_calls);
}
unexpected();
});

func (p *Parser) expr_list(closing, allow_trailing_comma, allow_empty) {
	var first = true, a = [];
	while(!is("punc", closing))
	{
		if (first) first = false;
		else expect(",");
		if (allow_trailing_comma && is("punc", closing))
		break;
		if (is("punc", ",") && allow_empty) {
			a.push([ "atom", "undefined" ]);
} else {
a.push(expression(false));
}
}
next();
return a;
};

func (p *Parser) array_() {
	return as("array", expr_list("]", !exigent_mode, true));
};

func (p *Parser) object_() {
	var first = true, a = [];
	while(!is("punc", "}"))
	{
		if (first) first = false;
		else expect(",");
		if (!exigent_mode && is("punc", "}"))
		// allow trailing comma
		break;
		var
		type = S.token.
		type;
		var name = as_property_name();
		if (
		type == "name" && (name == "get" || name == "set") && !is("punc", ":")) {
		a.push([ as_name(), function_(false), name ]);
} else {
expect(":");
a.push([ name, expression(false) ]);
}
}
next();
return as("object", a);
};

func (p *Parser) as_property_name() {
	switch
	(S.token.
	type) {
case "num":
case "string":
return prog1(S.token.value, next);
}
return as_name();
};

func (p *Parser) as_name() {
	switch
	(S.token.
	type) {
case "name":
case "operator":
case "keyword":
case "atom":
return prog1(S.token.value, next);
default:
unexpected();
}
};

func (p *Parser) subscripts(expr, allow_calls) {
	if (is("punc", ".")) {
		next();
		return subscripts(as("dot", expr, as_name()), allow_calls);
	}
	if (is("punc", "[")) {
		next();
		return subscripts(as("sub", expr, prog1(expression, curry(expect, "]"))), allow_calls);
	}
	if (allow_calls && is("punc", "(")) {
		next();
		return subscripts(as("call", expr, expr_list(")")), true);
	}
	return expr;
};

func (p *Parser) maybe_unary(allow_calls) {
	if (is("operator") && HOP(UNARY_PREFIX, S.token.value)) {
		return make_unary("unary-prefix",
			prog1(S.token.value, next),
			maybe_unary(allow_calls));
	}
	var val = expr_atom(allow_calls);
	while(is("operator") && HOP(UNARY_POSTFIX, S.token.value) && !S.token.nlb)
	{
		val = make_unary("unary-postfix", S.token.value, val);
		next();
	}
	return val;
};

func (p *Parser) make_unary(tag, op, expr) {
	if ((op == "++" || op == "--") && !is_assignable(expr))
		croak("Invalid use of " + op + " operator");
	return as(tag, op, expr);
};

func (p *Parser) expr_op(left, min_prec, no_in) {
	var op = is("operator") ? S.token.value : null;
	if (op && op == "in" && no_in) op = null;
	var prec = op != null ? PRECEDENCE[op] : null;
	if (prec != null && prec > min_prec) {
		next();
		var right = expr_op(maybe_unary(true), prec, no_in);
		return expr_op(as("binary", op, left, right), min_prec, no_in);
	}
	return left;
};

func (p *Parser) expr_ops(no_in) {
	return expr_op(maybe_unary(true), 0, no_in);
};

func (p *Parser) maybe_conditional(no_in) {
	var expr = expr_ops(no_in);
	if (is("operator", "?")) {
		next();
		var yes = expression(false);
		expect(":");
		return as("conditional", expr, yes, expression(false, no_in));
	}
	return expr;
};

func (p *Parser) is_assignable(expr) {
	if (!exigent_mode)
	return true;
	switch (expr[0] + "") {
	case "dot":
	case "sub":
	case "new":
	case "call":
		return true;
	case "name":
		return expr[1] != "this";
	}
};

func (p *Parser) maybe_assign(no_in) {
	var left = maybe_conditional(no_in), val = S.token.value;
	if (is("operator") && HOP(ASSIGNMENT, val)) {
		if (is_assignable(left)) {
			next();
			return as("assign", ASSIGNMENT[val], left, maybe_assign(no_in));
		}
		croak("Invalid assignment");
	}
	return left;
};

var expression = maybe_embed_tokens(function(commas, no_in) {
if (arguments.length == 0)
commas = true;

var expr = maybe_assign(no_in);
if (commas && is("punc", ",")) {
next();
return as("seq", expr, expression(true, no_in));
}
return expr;
});

func (p *Parser) in_loop(cont) {
	try{
		++S.in_loop;
		return cont();
	}
	finally{
		--S.in_loop;
	}
};
//
//return as("toplevel", (function(a){
//while (!is("eof"))
//a.push(statement());
//return a;
//})([]));
//
//};
