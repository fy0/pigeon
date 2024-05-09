// Code generated by static_code_generator with go generate; DO NOT EDIT.

package builder

var staticCode = `
var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEntrypoint is returned when the specified entrypoint rule
	// does not exit.
	errInvalidEntrypoint = errors.New("invalid entrypoint")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errMaxExprCnt is used to signal that the maximum number of
	// expressions have been parsed.
	errMaxExprCnt = errors.New("max number of expressions parsed")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// MaxExpressions creates an Option to stop parsing after the provided
// number of expressions have been parsed, if the value is 0 then the parser will
// parse for as many steps as needed (possibly an infinite number).
//
// The default for maxExprCnt is 0.
func MaxExpressions(maxExprCnt uint64) Option {
	return func(p *parser) Option {
		oldMaxExprCnt := p.maxExprCnt
		p.maxExprCnt = maxExprCnt
		return MaxExpressions(oldMaxExprCnt)
	}
}

// Entrypoint creates an Option to set the rule name to use as entrypoint.
// The rule name must have been specified in the -alternate-entrypoints
// if generating the parser with the -optimize-grammar flag, otherwise
// it may have been optimized out. Passing an empty string sets the
// entrypoint to the first rule in the grammar.
//
// The default is to start parsing at the first rule in the grammar.
func Entrypoint(ruleName string) Option {
	return func(p *parser) Option {
		oldEntrypoint := p.entrypoint
		p.entrypoint = ruleName
		if ruleName == "" {
			p.entrypoint = g.rules[0].name
		}
		return Entrypoint(oldEntrypoint)
	}
}

// ==template== {{ if not .Optimize }}
// Statistics adds a user provided Stats struct to the parser to allow
// the user to process the results after the parsing has finished.
// Also the key for the "no match" counter is set.
//
// Example usage:
//
//	input := "input"
//	stats := Stats{}
//	_, err := Parse("input-file", []byte(input), Statistics(&stats, "no match"))
//	if err != nil {
//	    log.Panicln(err)
//	}
//	b, err := json.MarshalIndent(stats.ChoiceAltCnt, "", "  ")
//	if err != nil {
//	    log.Panicln(err)
//	}
//	fmt.Println(string(b))
func Statistics(stats *Stats, choiceNoMatch string) Option {
	return func(p *parser) Option {
		oldStats := p.Stats
		p.Stats = stats
		oldChoiceNoMatch := p.choiceNoMatch
		p.choiceNoMatch = choiceNoMatch
		if p.Stats.ChoiceAltCnt == nil {
			p.Stats.ChoiceAltCnt = make(map[string]map[string]int)
		}
		return Statistics(oldStats, oldChoiceNoMatch)
	}
}

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// {{ end }} ==template==

// AllowInvalidUTF8 creates an Option to allow invalid UTF-8 bytes.
// Every invalid UTF-8 byte is treated as a utf8.RuneError (U+FFFD)
// by character class matchers and is matched by the any matcher.
// The returned matched value, c.text and c.offset are NOT affected.
//
// The default is false.
func AllowInvalidUTF8(b bool) Option {
	return func(p *parser) Option {
		old := p.allowInvalidUTF8
		p.allowInvalidUTF8 = b
		return AllowInvalidUTF8(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// GlobalStore creates an Option to set a key to a certain value in
// the globalStore.
func GlobalStore(key string, value any) Option {
	return func(p *parser) Option {
		old := p.cur.globalStore[key]
		p.cur.globalStore[key] = value
		return GlobalStore(key, old)
	}
}

// ==template== {{ if or .GlobalState (not .Optimize) }}

// InitState creates an Option to set a key to a certain value in
// the global "state" store.
func InitState(key string, value any) Option {
	return func(p *parser) Option {
		old := p.cur.state[key]
		p.cur.state[key] = value
		return InitState(key, old)
	}
}

// {{ end }} ==template==

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (i any, err error) { //{{ if .Nolint }} nolint: deadcode {{else}} ==template== {{ end }}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = closeErr
		}
	}()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (any, error) { //{{ if .Nolint }} nolint: deadcode {{else}} ==template== {{ end }}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (any, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return strconv.Itoa(p.line) + ":" + strconv.Itoa(p.col) + " [" + strconv.Itoa(p.offset) + "]"
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match

	// ==template== {{ if or .GlobalState (not .Optimize) }}

	// state is a store for arbitrary key,value pairs that the user wants to be
	// tied to the backtracking of the parser.
	// This is always rolled back if a parsing rule fails.
	state storeDict

	// {{ end }} ==template==

	// globalStore is a general store for the user to store arbitrary key-value
	// pairs that they need to manage and that they do not want tied to the
	// backtracking of the parser. This is only modified by the user and never
	// rolled back by the parser. It is always up to the user to keep this in a
	// consistent state.
	globalStore storeDict
}

type storeDict map[string]any

// the AST types...

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type grammar struct {
	pos   position
	rules []*rule
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type rule struct {
	pos         position
	name        string
	displayName string
	expr        any

	// ==template== {{ if .LeftRecursion }}
	leader        bool
	leftRecursive bool
	// {{ end }} ==template==
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type choiceExpr struct {
	pos          position
	alternatives []any
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type actionExpr struct {
	pos  position
	expr any
	run  func(*parser) (any, error)
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type recoveryExpr struct {
	pos          position
	expr         any
	recoverExpr  any
	failureLabel []string
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type seqExpr struct {
	pos   position
	exprs []any
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type throwExpr struct {
	pos   position
	label string
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type labeledExpr struct {
	pos   position
	label string
	expr  any
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type expr struct {
	pos  position
	expr any
}

type (
	andExpr        expr //{{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
	notExpr        expr //{{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
	zeroOrOneExpr  expr //{{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
	zeroOrMoreExpr expr //{{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
	oneOrMoreExpr  expr //{{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
)

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type ruleRefExpr struct {
	pos  position
	name string
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
	want       string
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type codeExpr struct {
	pos position
	run func(*parser) (any, error)
}

// {{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}
type charClassMatcher struct {
	pos             position
	val             string
	basicLatinChars [128]bool
	chars           []rune
	ranges          []rune
	classes         []*unicode.RangeTable
	ignoreCase      bool
	inverted        bool
}

type anyMatcher position //{{ if .Nolint }} nolint: structcheck {{else}} ==template== {{ end }}

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner    error
	pos      position
	prefix   string
	expected []string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	stats := Stats{
		ChoiceAltCnt: make(map[string]map[string]int),
	}

	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
		cur: current{
			// ==template== {{ if or .GlobalState (not .Optimize) }}
			state: make(storeDict),
			// {{ end }} ==template==
			globalStore: make(storeDict),
		},
		maxFailPos:      position{col: 1, line: 1},
		maxFailExpected: make([]string, 0, 20),
		Stats:           &stats,
		// start rule is rule [0] unless an alternate entrypoint is specified
		entrypoint: g.rules[0].name,
	}
	p.setOptions(opts)

	if p.maxExprCnt == 0 {
		p.maxExprCnt = math.MaxUint64
	}

	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

// {{ if .Nolint }} nolint: structcheck,deadcode {{else}} ==template== {{ end }}
type resultTuple struct {
	v   any
	b   bool
	end savepoint
}

// {{ if .Nolint }} nolint: varcheck {{else}} ==template== {{ end }}
const choiceNoMatch = -1

// Stats stores some statistics, gathered during parsing
type Stats struct {
	// ExprCnt counts the number of expressions processed during parsing
	// This value is compared to the maximum number of expressions allowed
	// (set by the MaxExpressions option).
	ExprCnt uint64

	// ChoiceAltCnt is used to count for each ordered choice expression,
	// which alternative is used how may times.
	// These numbers allow to optimize the order of the ordered choice expression
	// to increase the performance of the parser
	//
	// The outer key of ChoiceAltCnt is composed of the name of the rule as well
	// as the line and the column of the ordered choice.
	// The inner key of ChoiceAltCnt is the number (one-based) of the matching alternative.
	// For each alternative the number of matches are counted. If an ordered choice does not
	// match, a special counter is incremented. The name of this counter is set with
	// the parser option Statistics.
	// For an alternative to be included in ChoiceAltCnt, it has to match at least once.
	ChoiceAltCnt map[string]map[string]int
}

// ==template== {{ if .LeftRecursion }}
type ruleWithExpsStack struct {
	rule   *rule
	estack []any
}

// {{ end }} ==template==

// {{ if .Nolint }} nolint: structcheck,maligned {{else}} ==template== {{ end }}
type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	depth   int
	recover bool
	// ==template== {{ if not .Optimize }}
	debug bool

	memoize bool
	// {{ end }} ==template==
	// ==template== {{ if or .LeftRecursion (not .Optimize) }}
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[any]resultTuple
	// {{ end }} ==template==

	// rules table, maps the rule identifier to the rule node
	rules map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]any
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// parse fail
	maxFailPos            position
	maxFailExpected       []string
	maxFailInvertExpected bool

	// max number of expressions to be parsed
	maxExprCnt uint64
	// entrypoint for the parser
	entrypoint string

	allowInvalidUTF8 bool

	*Stats

	choiceNoMatch string
	// recovery expression stack, keeps track of the currently available recovery expression, these are traversed in reverse
	recoveryStack []map[string]any
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]any)
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

// push a recovery expression with its labels to the recoveryStack
func (p *parser) pushRecovery(labels []string, expr any) {
	if cap(p.recoveryStack) == len(p.recoveryStack) {
		// create new empty slot in the stack
		p.recoveryStack = append(p.recoveryStack, nil)
	} else {
		// slice to 1 more
		p.recoveryStack = p.recoveryStack[:len(p.recoveryStack)+1]
	}

	m := make(map[string]any, len(labels))
	for _, fl := range labels {
		m[fl] = expr
	}
	p.recoveryStack[len(p.recoveryStack)-1] = m
}

// pop a recovery expression from the recoveryStack
func (p *parser) popRecovery() {
	// GC that map
	p.recoveryStack[len(p.recoveryStack)-1] = nil

	p.recoveryStack = p.recoveryStack[:len(p.recoveryStack)-1]
}

// ==template== {{ if not .Optimize }}
func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) printIndent(mark string, s string) string {
	return p.print(strings.Repeat(" ", p.depth)+mark, s)
}

func (p *parser) in(s string) string {
	res := p.printIndent(">", s)
	p.depth++
	return res
}

func (p *parser) out(s string) string {
	p.depth--
	return p.printIndent("<", s)
}

// {{ end }} ==template==

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position, []string{})
}

func (p *parser) addErrAt(err error, pos position, expected []string) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String(), expected: expected}
	p.errs.add(pe)
}

func (p *parser) failAt(fail bool, pos position, want string) {
	// process fail if parsing fails and not inverted or parsing succeeds and invert is set
	if fail == p.maxFailInvertExpected {
		if pos.offset < p.maxFailPos.offset {
			return
		}

		if pos.offset > p.maxFailPos.offset {
			p.maxFailPos = pos
			p.maxFailExpected = p.maxFailExpected[:0]
		}

		if p.maxFailInvertExpected {
			want = "!" + want
		}
		p.maxFailExpected = append(p.maxFailExpected, want)
	}
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError && n == 1 { // see utf8.DecodeRune
		if !p.allowInvalidUTF8 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("restore"))
	}
	// {{ end }} ==template==
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// ==template== {{ if or .GlobalState (not .Optimize) }}

// Cloner is implemented by any value that has a Clone method, which returns a
// copy of the value. This is mainly used for types which are not passed by
// value (e.g map, slice, chan) or structs that contain such types.
//
// This is used in conjunction with the global state feature to create proper
// copies of the state to allow the parser to properly restore the state in
// the case of backtracking.
type Cloner interface {
	Clone() any
}

var statePool = &sync.Pool{
	New: func() any { return make(storeDict) },
}

func (sd storeDict) Discard() {
	for k := range sd {
		delete(sd, k)
	}
	statePool.Put(sd)
}

// clone and return parser current state.
func (p *parser) cloneState() storeDict {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("cloneState"))
	}
	// {{ end }} ==template==

	state := statePool.Get().(storeDict)
	for k, v := range p.cur.state {
		if c, ok := v.(Cloner); ok {
			state[k] = c.Clone()
		} else {
			state[k] = v
		}
	}
	return state
}

// restore parser current state to the state storeDict.
// every restoreState should applied only one time for every cloned state
func (p *parser) restoreState(state storeDict) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("restoreState"))
	}
	// {{ end }} ==template==
	p.cur.state.Discard()
	p.cur.state = state
}

// {{ end }} ==template==

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

// ==template== {{ if or .LeftRecursion (not .Optimize) }}
func (p *parser) getMemoized(node any) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node any, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[any]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[any]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

// {{ end }} ==template==

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

// {{ if .Nolint }} nolint: gocyclo {{else}} ==template== {{ end }}
func (p *parser) parse(g *grammar) (val any, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				// ==template== {{ if not .Optimize }}
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				// {{ end }} ==template==
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	startRule, ok := p.rules[p.entrypoint]
	if !ok {
		p.addErr(errInvalidEntrypoint)
		return nil, p.errs.err()
	}

	p.read() // advance to first rune
	val, ok = p.parseRuleWrap(startRule)
	if !ok {
		if len(*p.errs) == 0 {
			// If parsing fails, but no errors have been recorded, the expected values
			// for the farthest parser position are returned as error.
			maxFailExpectedMap := make(map[string]struct{}, len(p.maxFailExpected))
			for _, v := range p.maxFailExpected {
				maxFailExpectedMap[v] = struct{}{}
			}
			expected := make([]string, 0, len(maxFailExpectedMap))
			eof := false
			if _, ok := maxFailExpectedMap["!."]; ok {
				delete(maxFailExpectedMap, "!.")
				eof = true
			}
			for k := range maxFailExpectedMap {
				expected = append(expected, k)
			}
			sort.Strings(expected)
			if eof {
				expected = append(expected, "EOF")
			}
			p.addErrAt(errors.New("no match found, expected: "+listJoin(expected, ", ", "or")), p.maxFailPos, expected)
		}

		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func listJoin(list []string, sep string, lastSep string) string {
	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	default:
		return strings.Join(list[:len(list)-1], sep) + " " + lastSep + " " + list[len(list)-1]
	}
}

// ==template== {{ if .LeftRecursion }}

func (p *parser) parseRuleRecursiveLeader(rule *rule) (any, bool) {
	result, ok := p.getMemoized(rule)
	if ok {
		p.restore(result.end)
		return result.v, result.b
	}

	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("recursive " + rule.name))
	}
	// {{ end }} ==template==

	var (
		depth      = 0
		startMark  = p.pt
		lastResult = resultTuple{nil, false, startMark}
		lastErrors = *p.errs
	)

	for {
		// ==template== {{ if or .GlobalState (not .Optimize) }}
		lastState := p.cloneState()
		// {{ end }} ==template==
		p.setMemoized(startMark, rule, lastResult)
		val, ok := p.parseRule(rule)
		endMark := p.pt
		// ==template== {{ if not .Optimize }}
		if p.debug {
			p.printIndent("RECURSIVE", fmt.Sprintf(
				"Rule %s depth %d: %t -> %s",
				rule.name, depth, ok, string(p.sliceFrom(startMark))))
		}
		// {{ end }} ==template==
		if (!ok) || (endMark.offset <= lastResult.end.offset && depth != 0) {
			// ==template== {{ if or .GlobalState (not .Optimize) }}
			p.restoreState(lastState)
			// {{ end }} ==template==
			*p.errs = lastErrors
			break
		}
		lastResult = resultTuple{val, ok, endMark}
		lastErrors = *p.errs
		p.restore(startMark)
		depth++
	}

	p.restore(lastResult.end)
	p.setMemoized(startMark, rule, lastResult)
	return lastResult.v, lastResult.b
}

func (p *parser) parseRuleRecursiveNoLeader(rule *rule) (any, bool) {
	return p.parseRule(rule)
}

// {{ end }} ==template==

// ==template== {{ if not .Optimize }}
func (p *parser) parseRuleMemoize(rule *rule) (any, bool) {
	res, ok := p.getMemoized(rule)
	if ok {
		p.restore(res.end)
		return res.v, res.b
	}

	startMark := p.pt
	val, ok := p.parseRule(rule)
	p.setMemoized(startMark, rule, resultTuple{val, ok, p.pt})

	return val, ok
}

// {{ end }} ==template==

func (p *parser) parseRuleWrap(rule *rule) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}
	// {{ end }} ==template==
	var (
		val any
		ok  bool
		// ==template== {{ if not .Optimize }}
		startMark = p.pt
		// {{ end }} ==template==
	)

	// ==template== {{ if and .LeftRecursion (not .Optimize) }}
	if p.memoize || rule.leftRecursive {
		if rule.leader {
			val, ok = p.parseRuleRecursiveLeader(rule)
		} else if p.memoize && !rule.leftRecursive {
			val, ok = p.parseRuleMemoize(rule)
		} else {
			val, ok = p.parseRuleRecursiveNoLeader(rule)
		}
	} else {
		val, ok = p.parseRule(rule)
	}
	// {{ else if not .Optimize }}
	if p.memoize {
		val, ok = p.parseRuleMemoize(rule)
	} else {
		val, ok = p.parseRule(rule)
	}
	// {{ else if .LeftRecursion }}
	if rule.leftRecursive {
		if rule.leader {
			val, ok = p.parseRuleRecursiveLeader(rule)
		} else {
			val, ok = p.parseRuleRecursiveNoLeader(rule)
		}
	} else {
		val, ok = p.parseRule(rule)
	}
	// {{ else }}
	val, ok = p.parseRule(rule)
	// {{ end }} ==template==

	// ==template== {{ if not .Optimize }}
	if ok && p.debug {
		p.printIndent("MATCH", string(p.sliceFrom(startMark)))
	}
	// {{ end }} ==template==
	return val, ok
}

func (p *parser) parseRule(rule *rule) (any, bool) {
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExprWrap(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	return val, ok
}

func (p *parser) parseExprWrap(expr any) (any, bool) {
	// ==template== {{ if not .Optimize }}
	var pt savepoint

	// ==template== {{ if .LeftRecursion }}
	isLeftRecursion := p.rstack[len(p.rstack)-1].leftRecursive
	if p.memoize && !isLeftRecursion {
	// {{ else }}
	if p.memoize {
	// {{ end }} ==template==
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	// {{ end }} ==template==
	val, ok := p.parseExpr(expr)

	// ==template== {{ if not .Optimize }}
	// ==template== {{ if .LeftRecursion }}
	if p.memoize && !isLeftRecursion {
	// {{ else }}
	if p.memoize {
	// {{ end }} ==template==
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	// {{ end }} ==template==
	return val, ok
}

// {{ if .Nolint }} nolint: gocyclo {{else}} ==template== {{ end }}
func (p *parser) parseExpr(expr any) (any, bool) {
	p.ExprCnt++
	if p.ExprCnt > p.maxExprCnt {
		panic(errMaxExprCnt)
	}

	var val any
	var ok bool
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *codeExpr:
		val, ok = p.parseCodeExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *recoveryExpr:
		val, ok = p.parseRecoveryExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *throwExpr:
		val, ok = p.parseThrowExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	// {{ end }} ==template==
	start := p.pt
	val, ok := p.parseExprWrap(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		// ==template== {{ if or .GlobalState (not .Optimize) }}
		state := p.cloneState()
		// {{ end }} ==template==
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position, []string{})
		}
		// ==template== {{ if or .GlobalState (not .Optimize) }}
		p.restoreState(state)
		// {{ end }} ==template==

		val = actVal
	}
	// ==template== {{ if not .Optimize }}
	if ok && p.debug {
		p.printIndent("MATCH", string(p.sliceFrom(start)))
	}
	// {{ end }} ==template==
	return val, ok
}

func (p *parser) parseAndExpr(and *andExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	// {{ end }} ==template==
	pt := p.pt
	// ==template== {{ if or .GlobalState (not .Optimize) }}
	state := p.cloneState()
	// {{ end }} ==template==
	p.pushV()
	_, ok := p.parseExprWrap(and.expr)
	p.popV()
	// ==template== {{ if or .GlobalState (not .Optimize) }}
	p.restoreState(state)
	// {{ end }} ==template==
	p.restore(pt)

	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	// {{ end }} ==template==
	if p.pt.rn == utf8.RuneError && p.pt.w == 0 {
		// EOF - see utf8.DecodeRune
		p.failAt(false, p.pt.position, ".")
		return nil, false
	}
	start := p.pt
	p.read()
	p.failAt(true, start.position, ".")
	return p.sliceFrom(start), true
}

// {{ if .Nolint }} nolint: gocyclo {{else}} ==template== {{ end }}
func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	// {{ end }} ==template==
	cur := p.pt.rn
	start := p.pt

	// ==template== {{ if .BasicLatinLookupTable }}
	if cur < 128 {
		if chr.basicLatinChars[cur] != chr.inverted {
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
		p.failAt(false, start.position, chr.val)
		return nil, false
	}
	// {{ end }} ==template==

	// can't match EOF
	if cur == utf8.RuneError && p.pt.w == 0 { // see utf8.DecodeRune
		p.failAt(false, start.position, chr.val)
		return nil, false
	}

	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		p.failAt(true, start.position, chr.val)
		return p.sliceFrom(start), true
	}
	p.failAt(false, start.position, chr.val)
	return nil, false
}

// ==template== {{ if not .Optimize }}

func (p *parser) incChoiceAltCnt(ch *choiceExpr, altI int) {
	choiceIdent := fmt.Sprintf("%s %d:%d", p.rstack[len(p.rstack)-1].name, ch.pos.line, ch.pos.col)
	m := p.ChoiceAltCnt[choiceIdent]
	if m == nil {
		m = make(map[string]int)
		p.ChoiceAltCnt[choiceIdent] = m
	}
	// We increment altI by 1, so the keys do not start at 0
	alt := strconv.Itoa(altI + 1)
	if altI == choiceNoMatch {
		alt = p.choiceNoMatch
	}
	m[alt]++
}

// {{ end }} ==template==

func (p *parser) parseChoiceExpr(ch *choiceExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}
	// {{ end }} ==template==

	for altI, alt := range ch.alternatives {
		// dummy assignment to prevent compile error if optimized
		_ = altI

		// ==template== {{ if or .GlobalState (not .Optimize) }}
		state := p.cloneState()
		// {{ end }} ==template==

		p.pushV()
		val, ok := p.parseExprWrap(alt)
		p.popV()
		if ok {
			// ==template== {{ if not .Optimize }}
			p.incChoiceAltCnt(ch, altI)
			// {{ end }} ==template==
			return val, ok
		}
		// ==template== {{ if or .GlobalState (not .Optimize) }}
		p.restoreState(state)
		// {{ end }} ==template==
	}
	// ==template== {{ if not .Optimize }}
	p.incChoiceAltCnt(ch, choiceNoMatch)
	// {{ end }} ==template==
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	// {{ end }} ==template==
	p.pushV()
	val, ok := p.parseExprWrap(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseCodeExpr(code *codeExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseCodeExpr"))
	}

	// {{ end }} ==template==
	val, err := code.run(p)
	if err != nil {
		p.addErr(err)
		return nil, true
	}
	return val, true
}

func (p *parser) parseLitMatcher(lit *litMatcher) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	// {{ end }} ==template==
	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.failAt(false, start.position, lit.want)
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	p.failAt(true, start.position, lit.want)
	return p.sliceFrom(start), true
}

func (p *parser) parseNotExpr(not *notExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	// {{ end }} ==template==
	pt := p.pt
	// ==template== {{ if or .GlobalState (not .Optimize) }}
	state := p.cloneState()
	// {{ end }} ==template==
	p.pushV()
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	_, ok := p.parseExprWrap(not.expr)
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	p.popV()
	// ==template== {{ if or .GlobalState (not .Optimize) }}
	p.restoreState(state)
	// {{ end }} ==template==
	p.restore(pt)

	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	// {{ end }} ==template==
	var vals []any

	for {
		p.pushV()
		val, ok := p.parseExprWrap(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRecoveryExpr(recover *recoveryExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseRecoveryExpr (" + strings.Join(recover.failureLabel, ",") + ")"))
	}

	// {{ end }} ==template==

	p.pushRecovery(recover.failureLabel, recover.recoverExpr)
	val, ok := p.parseExprWrap(recover.expr)
	p.popRecovery()

	return val, ok
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	// {{ end }} ==template==
	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRuleWrap(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	// {{ end }} ==template==
	vals := make([]any, 0, len(seq.exprs))

	pt := p.pt
	// ==template== {{ if or .GlobalState (not .Optimize) }}
	state := p.cloneState()
	// {{ end }} ==template==
	for _, expr := range seq.exprs {
		val, ok := p.parseExprWrap(expr)
		if !ok {
			// ==template== {{ if or .GlobalState (not .Optimize) }}
			p.restoreState(state)
			// {{ end }} ==template==
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseThrowExpr(expr *throwExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseThrowExpr"))
	}

	// {{ end }} ==template==

	for i := len(p.recoveryStack) - 1; i >= 0; i-- {
		if recoverExpr, ok := p.recoveryStack[i][expr.label]; ok {
			if val, ok := p.parseExprWrap(recoverExpr); ok {
				return val, ok
			}
		}
	}

	return nil, false
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	// {{ end }} ==template==
	var vals []any

	for {
		p.pushV()
		val, ok := p.parseExprWrap(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (any, bool) {
	// ==template== {{ if not .Optimize }}
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	// {{ end }} ==template==
	p.pushV()
	val, _ := p.parseExprWrap(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}

`
