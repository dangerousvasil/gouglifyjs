package parse_js

/* -----[ Parser (constants) ]----- */

var UNARY_PREFIX = ArrayToHash([]string{
	"typeof",
	"void",
	"delete",
	"--",
	"++",
	"!",
	"~",
	"-",
	"+",
})

var UNARY_POSTFIX = ArrayToHash([]string{"--", "++"})

var ASSIGNMENT = (func(a []string, ret map[string]string, i int) map[string]string {
	for i < len(a) {
		ret[a[i]] = a[i][0 : len(a[i])-1]
		i++
	}
	return ret
})(
	[]string{"+=", "-=", "/=", "*=", "%=", ">>=", "<<=", ">>>=", "|=", "^=", "&="},
	map[string]string{"=": "true"},
	0,
)

var PRECEDENCE = (func(a [][]string, ret map[string]int) map[string]int {
	for n, v1 := range a {
		for _, v2 := range v1 {
			ret[v2] = n + 1
		}
	}
	return ret
})([][]string{
	{"||"},
	{"&&"},
	{"|"},
	{"^"},
	{"&"},
	{"==", "===", "!=", "!=="},
	{"<", ">", "<=", ">=", "in", "instanceof"},
	{">>", "<<", ">>>"},
	{"+", "-"},
	{"*", "/", "%"},
},
	map[string]int{},
)

var STATEMENTS_WITH_LABELS = ArrayToHash([]string{"for", "do", "while", "switch"})

var ATOMIC_START_TOKEN = ArrayToHash([]string{"atom", "num", "string", "regexp", "name"})
