package actionlint

import (
	"fmt"
	"strconv"
	"strings"
)

func errorAtToken(t *Token, msg string) *ExprError {
	return &ExprError{
		Message: msg,
		Offset:  t.Offset,
		Line:    t.Line,
		Column:  t.Column,
	}
}

// ExprParser is a parser for expression syntax. To know the details, see
// https://docs.github.com/en/actions/learn-github-actions/expressions
type ExprParser struct {
	cur   *Token
	lexer *ExprLexer
	err   *ExprError
}

// NewExprParser creates new ExprParser instance.
func NewExprParser() *ExprParser {
	return &ExprParser{}
}

func (p *ExprParser) error(msg string) {
	if p.err == nil {
		p.err = errorAtToken(p.cur, msg)
	}
}

func (p *ExprParser) errorf(format string, args ...interface{}) {
	p.error(fmt.Sprintf(format, args...))
}

func (p *ExprParser) unexpected(where string, expected []TokenKind) {
	if p.err != nil {
		return
	}
	qb := quotesBuilder{}
	for _, e := range expected {
		qb.append(e.String())
	}
	var what string
	if p.cur.Kind == TokenKindEnd {
		what = "end of input"
	} else {
		what = fmt.Sprintf("token %q", p.cur.Kind.String())
	}
	msg := fmt.Sprintf("unexpected %s while parsing %s. expecting %s", what, where, qb.build())
	p.error(msg)
}

func (p *ExprParser) next() *Token {
	ret := p.cur
	p.cur = p.lexer.Next()
	return ret
}

func (p *ExprParser) peek() *Token {
	return p.cur
}

func (p *ExprParser) parseIdent() ExprNode {
	ident := p.next() // eat ident
	switch p.peek().Kind {
	case TokenKindLeftParen:
		// Parse function call as primary expression though generally function call is parsed as
		// postfix expression. The reason is that only built-in function call is allowed in workflow
		// expression syntax, meant that callee is always built-in function name, not a general
		// expression.
		p.next() // eat '('
		args := []ExprNode{}
		if p.peek().Kind == TokenKindRightParen {
			// no arguments
			p.next() // eat ')'
		} else {
		LoopArgs:
			for {
				arg := p.parseLogicalOr()
				if arg == nil {
					return nil
				}

				args = append(args, arg)

				switch p.peek().Kind {
				case TokenKindComma:
					p.next() // eat ','
					// continue to next argument
				case TokenKindRightParen:
					p.next() // eat ')'
					break LoopArgs
				default:
					p.unexpected("arguments of function call", []TokenKind{TokenKindComma, TokenKindRightParen})
					return nil
				}
			}
		}
		return &FuncCallNode{ident.Value, args, ident}
	default:
		// Handle keywords. Note that keywords are case sensitive. TRUE, FALSE, NULL are invalid named value.
		switch ident.Value {
		case "null":
			return &NullNode{ident}
		case "true":
			return &BoolNode{true, ident}
		case "false":
			return &BoolNode{false, ident}
		default:
			// Variable name access is case insensitive. github.event and GITHUB.event are the same.
			return &VariableNode{strings.ToLower(ident.Value), ident}
		}
	}
}

func (p *ExprParser) parseNestedExpr() ExprNode {
	p.next() // eat '('

	nested := p.parseLogicalOr()
	if nested == nil {
		return nil
	}

	if p.peek().Kind == TokenKindRightParen {
		p.next() // eat ')'
	} else {
		p.unexpected("closing ')' of nexted expression (...)", []TokenKind{TokenKindRightParen})
		return nil
	}

	return nested
}

func (p *ExprParser) parseInt() ExprNode {
	t := p.peek()
	i, err := strconv.ParseInt(t.Value, 0, 32)
	if err != nil {
		p.errorf("parsing invalid integer literal %q: %s", t.Value, err)
		return nil
	}

	p.next() // eat int

	return &IntNode{int(i), t}
}

func (p *ExprParser) parseFloat() ExprNode {
	t := p.peek()
	f, err := strconv.ParseFloat(t.Value, 64)
	if err != nil {
		p.errorf("parsing invalid float literal %q: %s", t.Value, err)
		return nil
	}

	p.next() // eat float

	return &FloatNode{f, t}
}

func (p *ExprParser) parseString() ExprNode {
	t := p.next() // eat string
	s := t.Value
	s = s[1 : len(s)-1]                  // strip first and last single quotes
	s = strings.ReplaceAll(s, "''", "'") // unescape ''
	return &StringNode{s, t}
}

func (p *ExprParser) parsePrimaryExpr() ExprNode {
	switch p.peek().Kind {
	case TokenKindIdent:
		return p.parseIdent()
	case TokenKindLeftParen:
		return p.parseNestedExpr()
	case TokenKindInt:
		return p.parseInt()
	case TokenKindFloat:
		return p.parseFloat()
	case TokenKindString:
		return p.parseString()
	default:
		p.unexpected(
			"variable access, function call, null, bool, int, float or string",
			[]TokenKind{
				TokenKindIdent,
				TokenKindLeftParen,
				TokenKindInt,
				TokenKindFloat,
				TokenKindString,
			},
		)
		return nil
	}
}

func (p *ExprParser) parsePostfixOp() ExprNode {
	ret := p.parsePrimaryExpr()
	if ret == nil {
		return nil
	}

	for {
		switch p.peek().Kind {
		case TokenKindDot:
			p.next() // eat '.'
			switch p.peek().Kind {
			case TokenKindStar:
				p.next() // eat '*'
				ret = &ArrayDerefNode{ret}
			case TokenKindIdent:
				t := p.next() // eat 'b' of 'a.b'
				// Property name is case insensitive. github.event and github.EVENT are the same
				ret = &ObjectDerefNode{ret, strings.ToLower(t.Value)}
			default:
				p.unexpected(
					"object property dereference like 'a.b' or array element dereference like 'a.*'",
					[]TokenKind{TokenKindIdent, TokenKindStar},
				)
				return nil
			}
		case TokenKindLeftBracket:
			p.next() // eat '['
			idx := p.parseLogicalOr()
			if idx == nil {
				return nil
			}
			ret = &IndexAccessNode{ret, idx}
			if p.peek().Kind != TokenKindRightBracket {
				p.unexpected("closing bracket ']' for index access", []TokenKind{TokenKindRightBracket})
				return nil
			}
			p.next() // eat ']'
		default:
			return ret
		}
	}
}

func (p *ExprParser) parsePrefixOp() ExprNode {
	t := p.peek()
	if t.Kind != TokenKindNot {
		return p.parsePostfixOp()
	}
	p.next() // eat '!' token

	o := p.parsePrefixOp()
	if o == nil {
		return nil
	}

	return &NotOpNode{o, t}
}

func (p *ExprParser) parseCompareBinOp() ExprNode {
	l := p.parsePrefixOp()
	if l == nil {
		return nil
	}

	k := CompareOpNodeKindInvalid
	switch p.peek().Kind {
	case TokenKindLess:
		k = CompareOpNodeKindLess
	case TokenKindLessEq:
		k = CompareOpNodeKindLessEq
	case TokenKindGreater:
		k = CompareOpNodeKindGreater
	case TokenKindGreaterEq:
		k = CompareOpNodeKindGreaterEq
	case TokenKindEq:
		k = CompareOpNodeKindEq
	case TokenKindNotEq:
		k = CompareOpNodeKindNotEq
	default:
		return l
	}
	p.next() // eat the operator token

	r := p.parseCompareBinOp()
	if r == nil {
		return nil
	}

	return &CompareOpNode{k, l, r}
}

func (p *ExprParser) parseLogicalAnd() ExprNode {
	l := p.parseCompareBinOp()
	if l == nil {
		return nil
	}
	if p.peek().Kind != TokenKindAnd {
		return l
	}
	p.next() // eat &&
	r := p.parseLogicalAnd()
	if r == nil {
		return nil
	}
	return &LogicalOpNode{LogicalOpNodeKindAnd, l, r}
}

func (p *ExprParser) parseLogicalOr() ExprNode {
	l := p.parseLogicalAnd()
	if l == nil {
		return nil
	}
	if p.peek().Kind != TokenKindOr {
		return l
	}
	p.next() // eat ||
	r := p.parseLogicalOr()
	if r == nil {
		return nil
	}
	return &LogicalOpNode{LogicalOpNodeKindOr, l, r}
}

// Err returns an error which was caused while previous parsing.
func (p *ExprParser) Err() *ExprError {
	if err := p.lexer.Err(); err != nil {
		return err
	}
	return p.err
}

// Parse parses token sequence lexed by a given lexer into syntax tree.
func (p *ExprParser) Parse(l *ExprLexer) (ExprNode, *ExprError) {
	// Init
	p.err = nil
	p.lexer = l
	p.cur = l.Next()

	root := p.parseLogicalOr()
	if err := p.Err(); err != nil {
		return nil, err
	}

	if t := p.peek(); t.Kind != TokenKindEnd {
		// It did not reach the end of sequence
		qb := quotesBuilder{}
		qb.append(t.Kind.String())
		c := 1
		for {
			t := l.Next()
			if t.Kind == TokenKindEnd {
				break
			}
			qb.append(t.Kind.String())
			c++
		}
		p.errorf("parser did not reach end of input after parsing the expression. %d remaining token(s) in the input: %s", c, qb.build())
		return nil, p.err
	}

	return root, nil
}
