package expr

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"github.com/shopspring/decimal"
)

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	NUM     = "NUM"
	VAL     = "VAL"

	PLUS      = "+" //加
	MINUS     = "-" //减，负数
	ASTERISK  = "*" //乘
	SLASH     = "/" //除
	BACKSLASH = "|" //整除
	PERCENT   = "%" //取余
	POWER     = "^" //次方

	LVERT = "[" //绝对值-开始
	RVERT = "]" //绝对值-结束

	LPAREN = "(" //括号
	RPAREN = ")" //括号
)

const (
	_ int = iota
	LOWEST
	SUM     // +, -
	PRODUCT // *, /, %, |
	PREFIX  // -X
	POW     // ^
	CALL    // (X), |X|
)

var precedences = map[string]int{
	PLUS:      SUM,
	MINUS:     SUM,
	SLASH:     PRODUCT,
	BACKSLASH: PRODUCT,
	ASTERISK:  PRODUCT,
	PERCENT:   PRODUCT,
	POWER:     POW,
	LVERT:     CALL,
	LPAREN:    CALL,
}

func Calc(input string, val ...interface{}) float64 {
	lexer := NewLex(input)
	parser := NewParser(lexer)

	idx := 0
	for /*i*/ _, m := range val {
		v := reflect.ValueOf(m)
		t := v.Type()
		for t.Kind() == reflect.Ptr {
			v = v.Elem()
			t = v.Type()
		}
		if t.Kind() == reflect.Map {
			for _, k := range reflect.Indirect(v).MapKeys() {
				v1 := reflect.ValueOf(reflect.Indirect(v).MapIndex(k).Interface())
				t1 := v1.Type()
				for t1.Kind() == reflect.Ptr {
					v1 = v1.Elem()
					t1 = v1.Type()
				}
				if t1.Kind() == reflect.Struct {
					for n := 0; n < t1.NumField(); n++ {
						if v1.Field(n).CanInterface() {
							v2 := reflect.ValueOf(v1.Field(n).Interface())
							t2 := v2.Type()
							for t2.Kind() == reflect.Ptr {
								v2 = v2.Elem()
								t2 = v2.Type()
							}
							float_num, _ := strconv.ParseFloat(fmt.Sprint(v2.Interface()), 64)
							parser.v[t1.Field(n).Name] = float_num
						}
					}
				} else {
					float_num, _ := strconv.ParseFloat(fmt.Sprint(v1.Interface()), 64)
					parser.v[k.String()] = float_num
				}
			}
		} else if t.Kind() == reflect.Struct {
			for n := 0; n < t.NumField(); n++ {
				if v.Field(n).CanInterface() {
					v2 := reflect.ValueOf(v.Field(n).Interface())
					t2 := v2.Type()
					for t2.Kind() == reflect.Ptr {
						v2 = v2.Elem()
						t2 = v2.Type()
					}
					float_num, _ := strconv.ParseFloat(fmt.Sprint(v2.Interface()), 64)
					parser.v[t.Field(n).Name] = float_num
				}
			}
		} else {
			float_num, _ := strconv.ParseFloat(fmt.Sprint(v.Interface()), 64)
			parser.v[fmt.Sprintf("@%v", idx)] = float_num
			idx++
		}
	}
	//fmt.Println(parser.v)

	exp := parser.ParseExpression(LOWEST)
	fmt.Println(exp)
	f, exact := Eval(exp).Float64()
	if exact {
		return f
	}
	float_num, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'f', decimal.DivisionPrecision, 64), 64)
	return float_num
}

var zero = decimal.Zero

func Eval(exp Expression) decimal.Decimal {
	switch node := exp.(type) {
	case *FloatLiteralExpression:
		return node.Value
	case *PrefixExpression:
		rightV := Eval(node.Right)
		return evalPrefixExpression(node.Operator, rightV)
	case *InfixExpression:
		leftV := Eval(node.Left)
		rightV := Eval(node.Right)
		return evalInfixExpression(leftV, node.Operator, rightV)
	}

	return zero //decimal.Zero
}

func evalPrefixExpression(operator string, right decimal.Decimal) decimal.Decimal {
	if operator == "-" {
		return right.Neg()
	}
	if operator == "[" {
		return right.Abs()
	}
	return zero //decimal.Zero
}

func evalInfixExpression(left decimal.Decimal, operator string, right decimal.Decimal) decimal.Decimal {

	switch operator {
	case "+":
		return left.Add(right)
	case "-":
		return left.Sub(right)
	case "*":
		return left.Mul(right)
	case "/":
		if right.String() != "0" {
			return left.Div(right)
		} else {
			return zero //decimal.Zero
		}
	case "|":
		if right.String() != "0" {
			return left.Div(right).Truncate(0)
		} else {
			return zero //decimal.Zero
		}
	case "%":
		if right.String() != "0" {
			return left.Mod(right)
		} else {
			return zero //decimal.Zero
		}
	case "^":
		return left.Pow(right)
	default:
		return zero //decimal.Zero
	}
}

type Token struct {
	Type    string
	Literal string
}

func newToken(tokenType string, c byte) Token {
	return Token{
		Type:    tokenType,
		Literal: string(c),
	}
}

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func NewLex(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '(':
		tok = newToken(LPAREN, l.ch)
	case ')':
		tok = newToken(RPAREN, l.ch)
	case '+':
		tok = newToken(PLUS, l.ch)
	case '-':
		tok = newToken(MINUS, l.ch)
	case '*':
		tok = newToken(ASTERISK, l.ch)
	case '/':
		tok = newToken(SLASH, l.ch)
	case '|':
		tok = newToken(BACKSLASH, l.ch)
	case '%':
		tok = newToken(PERCENT, l.ch)
	case '^':
		tok = newToken(POWER, l.ch)
	case '[':
		tok = newToken(LVERT, l.ch)
	case ']':
		tok = newToken(RVERT, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if l.ch == '?' {
			tok.Type = VAL
			l.readChar()
			tok.Literal = "?"
			return tok
		} else if isDigit(l.ch) {
			tok.Type = NUM
			tok.Literal = l.readNumber()
			return tok
		} else if isMacroVal(l.ch) {
			tok.Type = VAL
			tok.Literal = l.readMacroVal()
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position]
}

func isMacroVal(ch byte) bool {
	return ('a' <= ch && ch <= 'z') ||
		('A' <= ch && ch <= 'Z') ||
		(ch == '_') || (ch > 127)
}

func (l *Lexer) readMacroVal() string {
	position := l.position
	for isMacroVal(l.ch) {
		l.readChar()
	}
	for isDigit(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// ast

type Expression interface {
	String() string
}

type FloatLiteralExpression struct {
	Token Token
	Value decimal.Decimal
}

func (il *FloatLiteralExpression) String() string { return il.Token.Literal }

type PrefixExpression struct {
	Token    Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	if pe.Operator == "[" {
		out.WriteString(pe.Operator)
		out.WriteString(pe.Right.String())
		out.WriteString("]")

	} else {
		out.WriteString("(")
		out.WriteString(pe.Operator)
		out.WriteString(pe.Right.String())
		out.WriteString(")")
	}

	return out.String()
}

type InfixExpression struct {
	Token    Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" ")
	out.WriteString(ie.Operator)
	out.WriteString(" ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

// parser
type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

type Parser struct {
	l     *Lexer
	v     map[string]float64
	index int

	curToken  Token
	peekToken Token

	prefixParseFns map[string]prefixParseFn

	errors []string
}

func (p *Parser) registerPrefix(tokenType string, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{
		l:      l,
		v:      map[string]float64{},
		index:  0,
		errors: []string{},
	}

	p.prefixParseFns = make(map[string]prefixParseFn)
	p.registerPrefix(NUM, p.parseFloatLiteral)
	p.registerPrefix(VAL, p.parseMacroValLiteral)
	p.registerPrefix(PLUS, p.parsePreNullExpression)
	p.registerPrefix(MINUS, p.parsePrefixExpression)
	p.registerPrefix(LVERT, p.parseAbsExpression)
	p.registerPrefix(LPAREN, p.parseGroupedExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	returnExp := prefix()

	for precedence < p.peekPrecedence() {
		p.nextToken()
		returnExp = p.parseInfixExpression(returnExp)
	}

	return returnExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekError(t string) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instend",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) expectPeek(t string) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekTokenIs(t string) bool {
	return p.peekToken.Type == t
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) parseFloatLiteral() Expression {

	lit := &FloatLiteralExpression{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = decimal.NewFromFloat(value)
	return lit
}

func (p *Parser) parseMacroValLiteral() Expression {

	lit := &FloatLiteralExpression{Token: p.curToken}

	value, has := p.v[lit.Token.Literal]
	if !has {
		value2, has := p.v[fmt.Sprintf("@%v", p.index)]
		if !has {
			msg := fmt.Sprintf("could not parse %q as macro value", p.curToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}
		if lit.Token.Literal != "?" {
			delete(p.v, fmt.Sprintf("@%v", p.index))
			p.v[lit.Token.Literal] = value2
		}
		value = value2
		p.index++
	}
	//fmt.Println(p.index, p.curToken, value)

	lit.Value = decimal.NewFromFloat(value)
	return lit
}

func (p *Parser) parsePrefixExpression() Expression {

	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.ParseExpression(PREFIX)
	return expression
}

func (p *Parser) parsePreNullExpression() Expression {
	p.nextToken()
	exp := p.ParseExpression(LOWEST)
	return exp
}

func (p *Parser) parseAbsExpression() Expression {
	exp := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	exp.Right = p.ParseExpression(LOWEST)

	if !p.expectPeek(RVERT) {
		return nil
	}
	return exp
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.ParseExpression(LOWEST)

	if !p.expectPeek(RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseInfixExpression(left Expression) Expression {

	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	expression.Right = p.ParseExpression(precedence)

	return expression
}

func (p *Parser) Errors() []string {
	return p.errors
}
