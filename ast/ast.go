// Package ast defines the abstract syntax tree for the PEG grammar.
//
// The parser generator's PEG grammar generates a tree using this package
// that is then converted by the builder to the simplified AST used in
// the generated parser.
package ast

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Pos represents a position in a source file.
type Pos struct {
	Filename string
	Line     int
	Col      int
	Off      int
}

// String returns the textual representation of a position.
func (p Pos) String() string {
	if p.Filename != "" {
		return fmt.Sprintf("%s:%d:%d (%d)", p.Filename, p.Line, p.Col, p.Off)
	}
	return fmt.Sprintf("%d:%d (%d)", p.Line, p.Col, p.Off)
}

// Grammar is the top-level node of the AST for the PEG grammar.
type Grammar struct {
	p     Pos
	Init  *CodeBlock
	Rules []*Rule
}

var _ Expression = (*Grammar)(nil)

// NewGrammar creates a new grammar at the specified position.
func NewGrammar(p Pos) *Grammar {
	return &Grammar{p: p}
}

// Pos returns the starting position of the node.
func (g *Grammar) Pos() Pos { return g.p }

// String returns the textual representation of a node.
func (g *Grammar) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s: %T{Init: %v, Rules: [\n",
		g.p, g, g.Init))
	for _, r := range g.Rules {
		buf.WriteString(fmt.Sprintf("%s,\n", r))
	}
	buf.WriteString("]}")
	return buf.String()
}

// NullableVisit recursively determines whether an object is nullable.
func (g *Grammar) NullableVisit(rules map[string]*Rule) bool {
	panic("NullableVisit should not be called on the Grammar")
}

// IsNullable returns the nullable attribute of the node.
func (g *Grammar) IsNullable() bool {
	panic("IsNullable should not be called on the Grammar")
}

// InitialNames returns names of nodes with which an expression can begin.
func (g *Grammar) InitialNames() map[string]struct{} {
	panic("InitialNames should not be called on the Grammar")
}

// Rule represents a rule in the PEG grammar. It has a name, an optional
// display name to be used in error messages, and an expression.
type Rule struct {
	p           Pos
	Name        *Identifier
	DisplayName *StringLit
	Expr        Expression

	IsLabelExists bool

	// Fields below to work with left recursion.
	Visited       bool
	Nullable      bool
	LeftRecursive bool
	Leader        bool
}

var _ Expression = (*Rule)(nil)

// NewRule creates a rule with at the specified position and with the
// specified name as identifier.
func NewRule(p Pos, name *Identifier) *Rule {
	return &Rule{p: p, Name: name}
}

// Pos returns the starting position of the node.
func (r *Rule) Pos() Pos { return r.p }

// String returns the textual representation of a node.
func (r *Rule) String() string {
	return fmt.Sprintf("%s: %T{Name: %v, DisplayName: %v, Expr: %v}",
		r.p, r, r.Name, r.DisplayName, r.Expr)
}

// NullableVisit recursively determines whether an object is nullable.
func (r *Rule) NullableVisit(rules map[string]*Rule) bool {
	if r.Visited {
		// A left-recursive rule is considered non-nullable.
		return false
	}
	r.Visited = true
	r.Nullable = r.Expr.NullableVisit(rules)
	r.Visited = false
	return r.Nullable
}

// IsNullable returns the nullable attribute of the node.
func (r *Rule) IsNullable() bool {
	return r.Nullable
}

// InitialNames returns names of nodes with which an expression can begin.
func (r *Rule) InitialNames() map[string]struct{} {
	return r.Expr.InitialNames()
}

// Expression is the interface implemented by all expression types.
type Expression interface {
	Pos() Pos

	// for work with left recursion
	NullableVisit(rules map[string]*Rule) bool
	IsNullable() bool
	InitialNames() map[string]struct{}
}

// ChoiceExpr is an ordered sequence of expressions. The parser tries to
// match any of the alternatives in sequence and stops at the first one
// that matches.
type ChoiceExpr struct {
	p            Pos
	Alternatives []Expression

	Nullable bool
}

var _ Expression = (*ChoiceExpr)(nil)

// NewChoiceExpr creates a choice expression at the specified position.
func NewChoiceExpr(p Pos) *ChoiceExpr {
	return &ChoiceExpr{p: p}
}

// Pos returns the starting position of the node.
func (c *ChoiceExpr) Pos() Pos { return c.p }

// String returns the textual representation of a node.
func (c *ChoiceExpr) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s: %T{Alternatives: [\n", c.p, c))
	for _, e := range c.Alternatives {
		buf.WriteString(fmt.Sprintf("%s,\n", e))
	}
	buf.WriteString("]}")
	return buf.String()
}

// NullableVisit recursively determines whether an object is nullable.
func (c *ChoiceExpr) NullableVisit(rules map[string]*Rule) bool {
	for _, alt := range c.Alternatives {
		if alt.NullableVisit(rules) {
			c.Nullable = true
			return true
		}
	}
	c.Nullable = false
	return false
}

// IsNullable returns the nullable attribute of the node.
func (c *ChoiceExpr) IsNullable() bool {
	return c.Nullable
}

// InitialNames returns names of nodes with which an expression can begin.
func (c *ChoiceExpr) InitialNames() map[string]struct{} {
	names := make(map[string]struct{})
	for _, alt := range c.Alternatives {
		for name := range alt.InitialNames() {
			names[name] = struct{}{}
		}
	}
	return names
}

// FailureLabel is an identifier, which can by thrown and recovered in a grammar.
type FailureLabel string

// RecoveryExpr is an ordered sequence of expressions. The parser tries to
// match any of the alternatives in sequence and stops at the first one
// that matches.
type RecoveryExpr struct {
	p           Pos
	Expr        Expression
	RecoverExpr Expression
	Labels      []FailureLabel

	Nullable bool
}

var _ Expression = (*RecoveryExpr)(nil)

// NewRecoveryExpr creates a choice expression at the specified position.
func NewRecoveryExpr(p Pos) *RecoveryExpr {
	return &RecoveryExpr{p: p}
}

// Pos returns the starting position of the node.
func (r *RecoveryExpr) Pos() Pos { return r.p }

// String returns the textual representation of a node.
func (r *RecoveryExpr) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s: %T{Expr: %v, RecoverExpr: %v", r.p, r, r.Expr, r.RecoverExpr))
	buf.WriteString(", Labels: [\n")
	for _, e := range r.Labels {
		buf.WriteString(fmt.Sprintf("%s,\n", e))
	}
	buf.WriteString("]}")
	return buf.String()
}

// NullableVisit recursively determines whether an object is nullable.
func (r *RecoveryExpr) NullableVisit(rules map[string]*Rule) bool {
	r.Nullable = r.Expr.NullableVisit(rules) || r.RecoverExpr.NullableVisit(rules)
	return r.Nullable
}

// IsNullable returns the nullable attribute of the node.
func (r *RecoveryExpr) IsNullable() bool {
	return r.Nullable
}

// InitialNames returns names of nodes with which an expression can begin.
func (r *RecoveryExpr) InitialNames() map[string]struct{} {
	names := make(map[string]struct{})
	for name := range r.Expr.InitialNames() {
		names[name] = struct{}{}
	}
	for name := range r.RecoverExpr.InitialNames() {
		names[name] = struct{}{}
	}
	return names
}

// ThrowExpr is an expression that throws an FailureLabel to be caught by a
// RecoveryChoiceExpr.
type ThrowExpr struct {
	p     Pos
	Label string
}

var _ Expression = (*ThrowExpr)(nil)

// NewThrowExpr creates a new throw expression at the specified position.
func NewThrowExpr(p Pos) *ThrowExpr {
	return &ThrowExpr{p: p}
}

// Pos returns the starting position of the node.
func (t *ThrowExpr) Pos() Pos { return t.p }

// String returns the textual representation of a node.
func (t *ThrowExpr) String() string {
	return fmt.Sprintf("%s: %T{Label: %v}", t.p, t, t.Label)
}

// NullableVisit recursively determines whether an object is nullable.
func (t *ThrowExpr) NullableVisit(rules map[string]*Rule) bool {
	return true
}

// IsNullable returns the nullable attribute of the node.
func (t *ThrowExpr) IsNullable() bool {
	return true
}

// InitialNames returns names of nodes with which an expression can begin.
func (t *ThrowExpr) InitialNames() map[string]struct{} {
	return make(map[string]struct{})
}

// SeqExpr is an ordered sequence of expressions, all of which must match
// if the SeqExpr is to be a match itself.
type SeqExpr struct {
	p     Pos
	Exprs []Expression

	Nullable bool
}

var _ Expression = (*SeqExpr)(nil)

// NewSeqExpr creates a new sequence expression at the specified position.
func NewSeqExpr(p Pos) *SeqExpr {
	return &SeqExpr{p: p}
}

// Pos returns the starting position of the node.
func (s *SeqExpr) Pos() Pos { return s.p }

// String returns the textual representation of a node.
func (s *SeqExpr) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s: %T{Exprs: [\n", s.p, s))
	for _, e := range s.Exprs {
		buf.WriteString(fmt.Sprintf("%s,\n", e))
	}
	buf.WriteString("]}")
	return buf.String()
}

// NullableVisit recursively determines whether an object is nullable.
func (s *SeqExpr) NullableVisit(rules map[string]*Rule) bool {
	for _, item := range s.Exprs {
		if !item.NullableVisit(rules) {
			s.Nullable = false
			return false
		}
	}
	s.Nullable = true
	return true
}

// IsNullable returns the nullable attribute of the node.
func (s *SeqExpr) IsNullable() bool {
	return s.Nullable
}

// InitialNames returns names of nodes with which an expression can begin.
func (s *SeqExpr) InitialNames() map[string]struct{} {
	names := make(map[string]struct{})
	for _, item := range s.Exprs {
		for name := range item.InitialNames() {
			names[name] = struct{}{}
		}
		if !item.IsNullable() {
			break
		}
	}
	return names
}

// LabeledExpr is an expression that has an associated label. Code blocks
// can access the value of the expression using that label, that becomes
// a local variable in the code.
type LabeledExpr struct {
	p           Pos
	Label       *Identifier
	Expr        Expression
	TextCapture bool
}

var _ Expression = (*LabeledExpr)(nil)

// NewLabeledExpr creates a new labeled expression at the specified position.
func NewLabeledExpr(p Pos) *LabeledExpr {
	return &LabeledExpr{p: p}
}

// Pos returns the starting position of the node.
func (l *LabeledExpr) Pos() Pos { return l.p }

// String returns the textual representation of a node.
func (l *LabeledExpr) String() string {
	return fmt.Sprintf("%s: %T{Label: %v, Expr: %v}", l.p, l, l.Label, l.Expr)
}

// NullableVisit recursively determines whether an object is nullable.
func (l *LabeledExpr) NullableVisit(rules map[string]*Rule) bool {
	return l.Expr.NullableVisit(rules)
}

// IsNullable returns the nullable attribute of the node.
func (l *LabeledExpr) IsNullable() bool {
	return l.Expr.IsNullable()
}

// InitialNames returns names of nodes with which an expression can begin.
func (l *LabeledExpr) InitialNames() map[string]struct{} {
	return l.Expr.InitialNames()
}

// AndExpr is a zero-length matcher that is considered a match if the
// expression it contains is a match.
// If Logical is true, return false when match empty string
type AndExpr struct {
	p       Pos
	Expr    Expression
	Logical bool
}

// NewAndExpr creates a new and (&) expression at the specified position.
func NewAndExpr(p Pos) *AndExpr {
	return &AndExpr{p: p}
}

var _ Expression = (*AndExpr)(nil)

// Pos returns the starting position of the node.
func (a *AndExpr) Pos() Pos { return a.p }

// String returns the textual representation of a node.
func (a *AndExpr) String() string {
	return fmt.Sprintf("%s: %T{Expr: %v}", a.p, a, a.Expr)
}

// NullableVisit recursively determines whether an object is nullable.
func (a *AndExpr) NullableVisit(rules map[string]*Rule) bool {
	return true
}

// IsNullable returns the nullable attribute of the node.
func (a *AndExpr) IsNullable() bool {
	return true
}

// InitialNames returns names of nodes with which an expression can begin.
func (a *AndExpr) InitialNames() map[string]struct{} {
	return make(map[string]struct{})
}

// NotExpr is a zero-length matcher that is considered a match if the
// expression it contains is not a match.
// If Logical is true, return false when match empty string
type NotExpr struct {
	p       Pos
	Expr    Expression
	Logical bool
}

var _ Expression = (*NotExpr)(nil)

// NewNotExpr creates a new not (!) expression at the specified position.
func NewNotExpr(p Pos) *NotExpr {
	return &NotExpr{p: p}
}

// Pos returns the starting position of the node.
func (n *NotExpr) Pos() Pos { return n.p }

// String returns the textual representation of a node.
func (n *NotExpr) String() string {
	return fmt.Sprintf("%s: %T{Expr: %v}", n.p, n, n.Expr)
}

// NullableVisit recursively determines whether an object is nullable.
func (n *NotExpr) NullableVisit(rules map[string]*Rule) bool {
	return true
}

// IsNullable returns the nullable attribute of the node.
func (n *NotExpr) IsNullable() bool {
	return true
}

// InitialNames returns names of nodes with which an expression can begin.
func (n *NotExpr) InitialNames() map[string]struct{} {
	return make(map[string]struct{})
}

// ZeroOrOneExpr is an expression that can be matched zero or one time.
type ZeroOrOneExpr struct {
	p    Pos
	Expr Expression
}

var _ Expression = (*ZeroOrOneExpr)(nil)

// NewZeroOrOneExpr creates a new zero or one expression at the specified
// position.
func NewZeroOrOneExpr(p Pos) *ZeroOrOneExpr {
	return &ZeroOrOneExpr{p: p}
}

// Pos returns the starting position of the node.
func (z *ZeroOrOneExpr) Pos() Pos { return z.p }

// String returns the textual representation of a node.
func (z *ZeroOrOneExpr) String() string {
	return fmt.Sprintf("%s: %T{Expr: %v}", z.p, z, z.Expr)
}

// NullableVisit recursively determines whether an object is nullable.
func (z *ZeroOrOneExpr) NullableVisit(rules map[string]*Rule) bool {
	return true
}

// IsNullable returns the nullable attribute of the node.
func (z *ZeroOrOneExpr) IsNullable() bool {
	return true
}

// InitialNames returns names of nodes with which an expression can begin.
func (z *ZeroOrOneExpr) InitialNames() map[string]struct{} {
	return z.Expr.InitialNames()
}

// ZeroOrMoreExpr is an expression that can be matched zero or more times.
type ZeroOrMoreExpr struct {
	p    Pos
	Expr Expression
}

var _ Expression = (*ZeroOrMoreExpr)(nil)

// NewZeroOrMoreExpr creates a new zero or more expression at the specified
// position.
func NewZeroOrMoreExpr(p Pos) *ZeroOrMoreExpr {
	return &ZeroOrMoreExpr{p: p}
}

// Pos returns the starting position of the node.
func (z *ZeroOrMoreExpr) Pos() Pos { return z.p }

// String returns the textual representation of a node.
func (z *ZeroOrMoreExpr) String() string {
	return fmt.Sprintf("%s: %T{Expr: %v}", z.p, z, z.Expr)
}

// NullableVisit recursively determines whether an object is nullable.
func (z *ZeroOrMoreExpr) NullableVisit(rules map[string]*Rule) bool {
	return true
}

// IsNullable returns the nullable attribute of the node.
func (z *ZeroOrMoreExpr) IsNullable() bool {
	return true
}

// InitialNames returns names of nodes with which an expression can begin.
func (z *ZeroOrMoreExpr) InitialNames() map[string]struct{} {
	return z.Expr.InitialNames()
}

// OneOrMoreExpr is an expression that can be matched one or more times.
type OneOrMoreExpr struct {
	p    Pos
	Expr Expression
}

var _ Expression = (*OneOrMoreExpr)(nil)

// NewOneOrMoreExpr creates a new one or more expression at the specified
// position.
func NewOneOrMoreExpr(p Pos) *OneOrMoreExpr {
	return &OneOrMoreExpr{p: p}
}

// Pos returns the starting position of the node.
func (o *OneOrMoreExpr) Pos() Pos { return o.p }

// String returns the textual representation of a node.
func (o *OneOrMoreExpr) String() string {
	return fmt.Sprintf("%s: %T{Expr: %v}", o.p, o, o.Expr)
}

// NullableVisit recursively determines whether an object is nullable.
func (o *OneOrMoreExpr) NullableVisit(rules map[string]*Rule) bool {
	return false
}

// IsNullable returns the nullable attribute of the node.
func (o *OneOrMoreExpr) IsNullable() bool {
	return false
}

// InitialNames returns names of nodes with which an expression can begin.
func (o *OneOrMoreExpr) InitialNames() map[string]struct{} {
	return o.Expr.InitialNames()
}

// RuleRefExpr is an expression that references a rule by name.
type RuleRefExpr struct {
	p    Pos
	Name *Identifier

	Nullable bool
}

var _ Expression = (*RuleRefExpr)(nil)

// NewRuleRefExpr creates a new rule reference expression at the specified
// position.
func NewRuleRefExpr(p Pos) *RuleRefExpr {
	return &RuleRefExpr{p: p}
}

// Pos returns the starting position of the node.
func (r *RuleRefExpr) Pos() Pos { return r.p }

// String returns the textual representation of a node.
func (r *RuleRefExpr) String() string {
	return fmt.Sprintf("%s: %T{Name: %v}", r.p, r, r.Name)
}

// NullableVisit recursively determines whether an object is nullable.
func (r *RuleRefExpr) NullableVisit(rules map[string]*Rule) bool {
	item, ok := rules[r.Name.Val]
	if !ok {
		// Token or unknown; never empty.
		r.Nullable = false
		return false
	}
	r.Nullable = item.NullableVisit(rules)
	return r.Nullable
}

// IsNullable returns the nullable attribute of the node.
func (r *RuleRefExpr) IsNullable() bool {
	return r.Nullable
}

// InitialNames returns names of nodes with which an expression can begin.
func (r *RuleRefExpr) InitialNames() map[string]struct{} {
	return map[string]struct{}{r.Name.Val: {}}
}

// StateCodeExpr is an expression which can modify the internal state of the parser.
type StateCodeExpr struct {
	p      Pos
	Code   *CodeBlock
	FuncIx int
}

// LitMatcher is a string literal matcher. The value to match may be a
// double-quoted string, a single-quoted single character, or a back-tick
// quoted raw string.
type LitMatcher struct {
	posValue   // can be str, rstr or char
	IgnoreCase bool
}

var _ Expression = (*LitMatcher)(nil)

// NewLitMatcher creates a new literal matcher at the specified position and
// with the specified value.
func NewLitMatcher(p Pos, v string) *LitMatcher {
	return &LitMatcher{posValue: posValue{p: p, Val: v}}
}

// Pos returns the starting position of the node.
func (l *LitMatcher) Pos() Pos { return l.p }

// String returns the textual representation of a node.
func (l *LitMatcher) String() string {
	return fmt.Sprintf("%s: %T{Val: %q, IgnoreCase: %t}", l.p, l, l.Val, l.IgnoreCase)
}

// NullableVisit recursively determines whether an object is nullable.
func (l *LitMatcher) NullableVisit(rules map[string]*Rule) bool {
	return l.IsNullable()
}

// IsNullable returns the nullable attribute of the node.
func (l *LitMatcher) IsNullable() bool {
	// The string token '' is considered empty.
	return len(l.Val) == 0
}

// InitialNames returns names of nodes with which an expression can begin.
func (l *LitMatcher) InitialNames() map[string]struct{} {
	return make(map[string]struct{})
}

// CharClassMatcher is a character class matcher. The value to match must
// be one of the specified characters, in a range of characters, or in the
// Unicode classes of characters.
type CharClassMatcher struct {
	posValue
	IgnoreCase     bool
	Inverted       bool
	Chars          []rune
	Ranges         []rune // pairs of low/high range
	UnicodeClasses []string
}

var _ Expression = (*CharClassMatcher)(nil)

// NewCharClassMatcher creates a new character class matcher at the specified
// position and with the specified raw value. It parses the raw value into
// the list of characters, ranges and Unicode classes.
func NewCharClassMatcher(p Pos, raw string) *CharClassMatcher {
	c := &CharClassMatcher{posValue: posValue{p: p, Val: raw}}
	c.parse()
	return c
}

func (c *CharClassMatcher) parse() {
	raw := c.Val
	c.IgnoreCase = strings.HasSuffix(raw, "i")
	if c.IgnoreCase {
		raw = raw[:len(raw)-1]
	}

	// "unquote" the character classes
	raw = raw[1 : len(raw)-1]
	if len(raw) == 0 {
		return
	}

	c.Inverted = raw[0] == '^'
	if c.Inverted {
		raw = raw[1:]
		if len(raw) == 0 {
			return
		}
	}

	// content of char class is necessarily valid, so escapes are correct
	r := strings.NewReader(raw)
	var chars []rune
	var buf bytes.Buffer
outer:
	for {
		rn, _, err := r.ReadRune()
		if err != nil {
			break outer
		}

		consumeN := 0
		switch rn {
		case '\\':
			rn, _, _ := r.ReadRune()
			switch rn {
			case ']':
				chars = append(chars, rn)
				continue

			case 'p':
				rn, _, _ := r.ReadRune()
				if rn == '{' {
					buf.Reset()
					for {
						rn, _, _ := r.ReadRune()
						if rn == '}' {
							break
						}
						buf.WriteRune(rn)
					}
					c.UnicodeClasses = append(c.UnicodeClasses, buf.String())
				} else {
					c.UnicodeClasses = append(c.UnicodeClasses, string(rn))
				}
				continue

			case 'x':
				consumeN = 2
			case 'u':
				consumeN = 4
			case 'U':
				consumeN = 8
			case '0', '1', '2', '3', '4', '5', '6', '7':
				consumeN = 2
			}

			buf.Reset()
			buf.WriteRune(rn)
			for i := 0; i < consumeN; i++ {
				rn, _, _ := r.ReadRune()
				buf.WriteRune(rn)
			}
			rn, _, _, _ = strconv.UnquoteChar("\\"+buf.String(), 0)
			chars = append(chars, rn)

		default:
			chars = append(chars, rn)
		}
	}

	// extract ranges and chars
	inRange, wasRange := false, false
	for i, r := range chars {
		if inRange {
			c.Ranges = append(c.Ranges, r)
			inRange = false
			wasRange = true
			continue
		}

		if r == '-' && !wasRange && len(c.Chars) > 0 && i < len(chars)-1 {
			inRange = true
			wasRange = false
			// start of range is the last Char added
			c.Ranges = append(c.Ranges, c.Chars[len(c.Chars)-1])
			c.Chars = c.Chars[:len(c.Chars)-1]
			continue
		}
		wasRange = false
		c.Chars = append(c.Chars, r)
	}
}

// Pos returns the starting position of the node.
func (c *CharClassMatcher) Pos() Pos { return c.p }

// String returns the textual representation of a node.
func (c *CharClassMatcher) String() string {
	return fmt.Sprintf("%s: %T{Val: %q, IgnoreCase: %t, Inverted: %t}",
		c.p, c, c.Val, c.IgnoreCase, c.Inverted)
}

// NullableVisit recursively determines whether an object is nullable.
func (c *CharClassMatcher) NullableVisit(rules map[string]*Rule) bool {
	return c.IsNullable()
}

// IsNullable returns the nullable attribute of the node.
func (c *CharClassMatcher) IsNullable() bool {
	return len(c.Chars) == 0 && len(c.Ranges) == 0 && len(c.UnicodeClasses) == 0
}

// InitialNames returns names of nodes with which an expression can begin.
func (c *CharClassMatcher) InitialNames() map[string]struct{} {
	return make(map[string]struct{})
}

// AnyMatcher is a matcher that matches any character except end-of-file.
type AnyMatcher struct {
	posValue
}

var _ Expression = (*AnyMatcher)(nil)

// NewAnyMatcher creates a new any matcher at the specified position. The
// value is provided for completeness' sake, but it is always the dot.
func NewAnyMatcher(p Pos, v string) *AnyMatcher {
	return &AnyMatcher{posValue{p, v}}
}

// Pos returns the starting position of the node.
func (a *AnyMatcher) Pos() Pos { return a.p }

// String returns the textual representation of a node.
func (a *AnyMatcher) String() string {
	return fmt.Sprintf("%s: %T{Val: %q}", a.p, a, a.Val)
}

// NullableVisit recursively determines whether an object is nullable.
func (a *AnyMatcher) NullableVisit(rules map[string]*Rule) bool {
	return false
}

// IsNullable returns the nullable attribute of the node.
func (a *AnyMatcher) IsNullable() bool {
	return false
}

// InitialNames returns names of nodes with which an expression can begin.
func (a *AnyMatcher) InitialNames() map[string]struct{} {
	return make(map[string]struct{})
}

// CodeBlock represents a code block.
type CodeBlock struct {
	posValue
}

var _ Expression = (*CodeBlock)(nil)

// NewCodeBlock creates a new code block at the specified position and with
// the specified value. The value includes the outer braces.
func NewCodeBlock(p Pos, code string) *CodeBlock {
	return &CodeBlock{posValue{p, code}}
}

// Pos returns the starting position of the node.
func (c *CodeBlock) Pos() Pos { return c.p }

// String returns the textual representation of a node.
func (c *CodeBlock) String() string {
	return fmt.Sprintf("%s: %T{Val: %q}", c.p, c, c.Val)
}

// NullableVisit recursively determines whether an object is nullable.
func (c *CodeBlock) NullableVisit(rules map[string]*Rule) bool {
	panic("NullableVisit should not be called on the CodeBlock")
}

// IsNullable returns the nullable attribute of the node.
func (c *CodeBlock) IsNullable() bool {
	panic("IsNullable should not be called on the CodeBlock")
}

// InitialNames returns names of nodes with which an expression can begin.
func (c *CodeBlock) InitialNames() map[string]struct{} {
	panic("InitialNames should not be called on the CodeBlock")
}

// Identifier represents an identifier.
type Identifier struct {
	posValue
}

var _ Expression = (*Identifier)(nil)

// NewIdentifier creates a new identifier at the specified position and
// with the specified name.
func NewIdentifier(p Pos, name string) *Identifier {
	return &Identifier{posValue{p: p, Val: name}}
}

// Pos returns the starting position of the node.
func (i *Identifier) Pos() Pos { return i.p }

// String returns the textual representation of a node.
func (i *Identifier) String() string {
	return fmt.Sprintf("%s: %T{Val: %q}", i.p, i, i.Val)
}

// NullableVisit recursively determines whether an object is nullable.
func (i *Identifier) NullableVisit(rules map[string]*Rule) bool {
	panic("NullableVisit should not be called on the Identifier")
}

// IsNullable returns the nullable attribute of the node.
func (i *Identifier) IsNullable() bool {
	panic("IsNullable should not be called on the Identifier")
}

// InitialNames returns names of nodes with which an expression can begin.
func (i *Identifier) InitialNames() map[string]struct{} {
	panic("InitialNames should not be called on the Identifier")
}

// StringLit represents a string literal.
type StringLit struct {
	posValue
}

var _ Expression = (*StringLit)(nil)

// NewStringLit creates a new string literal at the specified position and
// with the specified value.
func NewStringLit(p Pos, val string) *StringLit {
	return &StringLit{posValue{p: p, Val: val}}
}

// Pos returns the starting position of the node.
func (s *StringLit) Pos() Pos { return s.p }

// String returns the textual representation of a node.
func (s *StringLit) String() string {
	return fmt.Sprintf("%s: %T{Val: %q}", s.p, s, s.Val)
}

// NullableVisit recursively determines whether an object is nullable.
func (s *StringLit) NullableVisit(rules map[string]*Rule) bool {
	panic("NullableVisit should not be called on the StringLit")
}

// IsNullable returns the nullable attribute of the node.
func (s *StringLit) IsNullable() bool {
	panic("IsNullable should not be called on the StringLit")
}

// InitialNames returns names of nodes with which an expression can begin.
func (s *StringLit) InitialNames() map[string]struct{} {
	panic("InitialNames should not be called on the StringLit")
}

type posValue struct {
	p   Pos
	Val string
}
