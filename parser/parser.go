package parser

import (
	"fmt"
	"strconv"

	"mk/ast"
	"mk/lexer"
	"mk/token"
)

const (
	_           int = iota
	LOWEST          // 执行最低有限级(即左绑定和右绑定能力最弱)
	EQUALS          // ==
	LESSGREATER     // > or <
	SUM             // +
	PRODUCT         // *
	PREFIX          // -X or !X
	CALL            // myFunction(X)
	INDEX           // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression               // 前缀表达式(!, -)
	infixParseFn  func(ast.Expression) ast.Expression // 中缀表达式(+,-,*,/...)
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// 注册前缀表达式的解析函数
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)         //标识符
	p.registerPrefix(token.INT, p.parseIntegerLiteral)       //数值
	p.registerPrefix(token.BANG, p.parsePrefixExpression)    //!
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)   //-(取负)
	p.registerPrefix(token.TRUE, p.parseBoolean)             //true
	p.registerPrefix(token.FALSE, p.parseBoolean)            //false
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression) //(
	p.registerPrefix(token.IF, p.parseIfExpression)          //if
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral) //function
	p.registerPrefix(token.STRING, p.parseStringLiteral)     //字符串
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)    //数组
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	// 注册中缀表达式的解析函数
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)     //'+'
	p.registerInfix(token.MINUS, p.parseInfixExpression)    //'-'(减)
	p.registerInfix(token.SLASH, p.parseInfixExpression)    //'/'(除)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression) //'*'
	p.registerInfix(token.EQ, p.parseInfixExpression)       //'='
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)   //'!='
	p.registerInfix(token.LT, p.parseInfixExpression)       //'<'
	p.registerInfix(token.GT, p.parseInfixExpression)       //'>'
	p.registerInfix(token.LPAREN, p.parseCallExpression)    //'('
	p.registerInfix(token.LBRACKET, p.parseIndexExpression) //数组下标表达式

	// 初始化:
	// 执行两遍nextToken()
	// 确保curToken和peekToken都已设置
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// 解析语法树入口
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t,
		p.peekToken.Type)

	p.errors = append(p.errors, msg)
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// 检查语句的类型
// 再调用解析具体语句类型的方法
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// 解析let类型语句
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// 模式:let x = .... 中
	// let 后面必须为标识符(token.IDENT, 比如x)
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// curToken 为 标识符时, peekToken必须为等于号(token.ASSIGN)
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	// 以最低优先级解析表达式
	stmt.Value = p.parseExpression(LOWEST)

	// 直到分号结束
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// 解析return类型语句
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// 以最低优先级解析表达式
	stmt.ReturnValue = p.parseExpression(LOWEST)

	// 直到分号结束
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// 解析表达式类型语句
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	// 以最低优先级解析表达式
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// 解析表达式
func (p *Parser) parseExpression(precedence int) ast.Expression {

	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// 解析int类型字面量
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// 解析中缀类型表达式
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	if expression.Operator == "+" {
		expression.Right = p.parseExpression(precedence - 1)
	} else {
		expression.Right = p.parseExpression(precedence)
	}

	return expression
}

// 解析前缀类型表达式
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	// 带入PREFIX的优先级解析后面的表达式
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// 下一个token的优先级
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// 当前token的优先级
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// 检查当前token的类型是否匹配
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// 检查下一个token的类型是否匹配
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// 检查 true / false 表达式
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// 检查 '( xxx )' 类型表达式
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

// 检查 'if (a) { b } else { c }' 类型表达式
func (p *Parser) parseIfExpression() ast.Expression {
	// IF 类型token
	expression := &ast.IfExpression{Token: p.curToken}

	// 期望'('
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	// 读取表达式
	expression.Condition = p.parseExpression(LOWEST)

	// 期望')'
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 期望'{'
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// 解析语句
	expression.Consequence = p.parseBlockStatement()

	// 如果有'ELSE'
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		// 解析else里面的语句
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	// ELSE 类型token
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	// 检查是否遇到 '}'
	for !p.curTokenIs(token.RBRACE) {
		stmt := p.parseStatement()

		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// 解析函数
func (p *Parser) parseFunctionLiteral() ast.Expression {
	// 'FUNCTION' token
	lit := &ast.FunctionLiteral{Token: p.curToken}

	// 期望'('
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// 解析函数参数
	lit.Parameters = p.parseFunctionParameters()

	// 期望 '{'
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// 解析方法体(语句列表)
	lit.Body = p.parseBlockStatement()
	return lit
}

// 解析函数参数
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	// 参数列表就是逗号间隔的标识符列表
	identifiers := []*ast.Identifier{}

	// 期望'('
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	// 解析标识符
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// 循环解析其他标识符
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// 期望 ')' 结束
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// 解析函数调用
// 例如: fn(x, y) { return x + y;} (1, 2);
//       或者使用之前定义好的参数: add(1, 2);
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	// 函数调用标识符 '('
	exp := &ast.CallExpression{Token: p.curToken, Function: function}

	// 解析函数调用参数
	// 函数参数为表达式列表
	// 例如: add(1+2, 3+4);
	exp.Arguments = p.parseExpressionList(token.RPAREN)

	return exp
}

// 解析函数调用参数
func (p *Parser) parseCallArguments() []ast.Expression {
	// 参数列表就是表达式列表
	args := []ast.Expression{}

	// 期望')'进行结束
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()

	// 解析调用参数
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

// 解析字符串字面量
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// 解析数组字面量
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// 解析数组类的表达式列表
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	// 直接碰到']'为空数组，直接结束
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()

	//
	list = append(list, p.parseExpression(LOWEST))

	// 每读到一个','代表数组里面的一个表达式
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	// 没有正确解析则为nil
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

// 解析下标
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {

	// left为数组/map
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()

	// '[]'之间解析出来为下标值
	exp.Index = p.parseExpression(LOWEST)

	// 碰到']'结束
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

// 解析数组字面量
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}
