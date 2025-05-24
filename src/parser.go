package main

import "fmt"

// Instr

type NodeInstr interface {
	nodeInstr()
	String() string
}

type NodeInstrAssign struct {
	ident string
	expr  NodeExpr
}
type NodeInstrIf struct {
	rel   NodeRel
	instr NodeInstr
}
type NodeInstrGoto struct{ label string }
type NodeInstrOutput struct{ term NodeTerm }
type NodeInstrLabel struct{ name string }

func (NodeInstrAssign) nodeInstr() {}
func (NodeInstrIf) nodeInstr()     {}
func (NodeInstrGoto) nodeInstr()   {}
func (NodeInstrOutput) nodeInstr() {}
func (NodeInstrLabel) nodeInstr()  {}

func (instr NodeInstrAssign) String() string {
	return "assign(" + instr.ident + " = " + instr.expr.String() + ")"
}
func (instr NodeInstrIf) String() string {
	return "if(" + instr.rel.String() + " -> " + instr.instr.String() + ")"
}
func (instr NodeInstrGoto) String() string {
	return "goto(" + instr.label + ")"
}
func (instr NodeInstrOutput) String() string {
	return "output(" + instr.term.String() + ")"
}
func (instr NodeInstrLabel) String() string {
	return "label(" + instr.name + ")"
}

// Expr

type NodeExpr interface {
	nodeExpr()
	String() string
}

type NodeExprTermBinop struct {
	lhs NodeTerm
	rhs NodeTerm
}
type NodeExprSingle struct{ term NodeTerm }
type NodeExprPlus NodeExprTermBinop
type NodeExprMinus NodeExprTermBinop
type NodeExprMul NodeExprTermBinop

func (NodeExprSingle) nodeExpr() {}
func (NodeExprPlus) nodeExpr()   {}
func (NodeExprMinus) nodeExpr()  {}
func (NodeExprMul) nodeExpr()    {}

func (expr NodeExprSingle) String() string {
	return "expr(single(" + expr.term.String() + "))"
}
func (expr NodeExprPlus) String() string {
	return "expr(plus(" + expr.lhs.String() + " + " + expr.rhs.String() + "))"
}
func (expr NodeExprMinus) String() string {
	return "expr(minus(" + expr.lhs.String() + " - " + expr.rhs.String() + "))"
}
func (expr NodeExprMul) String() string {
	return "expr(mul(" + expr.lhs.String() + " * " + expr.rhs.String() + "))"
}

// Rel

type NodeRel interface {
	nodeRel()
	String() string
}

type NodeRelTermBinop struct {
	lhs NodeTerm
	rhs NodeTerm
}
type NodeRelLessThan NodeRelTermBinop
type NodeRelLessThanEqual NodeRelTermBinop
type NodeRelEqual NodeRelTermBinop

func (NodeRelLessThan) nodeRel() {}
func (rel NodeRelLessThan) String() string {
	return "rel(" + rel.lhs.String() + " < " + rel.rhs.String() + ")"
}
func (NodeRelLessThanEqual) nodeRel() {}
func (rel NodeRelLessThanEqual) String() string {
	return "rel(" + rel.lhs.String() + " <= " + rel.rhs.String() + ")"
}
func (NodeRelEqual) nodeRel() {}
func (rel NodeRelEqual) String() string {
	return "rel(" + rel.lhs.String() + " == " + rel.rhs.String() + ")"
}

// Term

type NodeTerm interface {
	nodeTerm()
	String() string
}

type NodeTermInput struct{}
type NodeTermInt struct{ val string }
type NodeTermIdent struct{ val string }

func (NodeTermInput) nodeTerm() {}
func (NodeTermInt) nodeTerm()   {}
func (NodeTermIdent) nodeTerm() {}

func (term NodeTermInput) String() string {
	return "term(input)"
}
func (term NodeTermInt) String() string {
	return "term(int(" + term.val + "))"
}
func (term NodeTermIdent) String() string {
	return "term(ident(" + term.val + "))"
}

// Over

type NodeProgram struct {
	instrs []NodeInstr
}

type Parser struct {
	ts    []Token
	index int
}

func (p *Parser) current(t *Token) {
	*t = p.ts[p.index]
}

func (p *Parser) advance() {
	p.index++
}

func createParser(ts []Token) Parser {
	var p Parser
	p.ts = ts

	return p
}

func (p *Parser) parseProgram() NodeProgram {
	var program NodeProgram
	program.instrs = make([]NodeInstr, 0)

	var t Token
	for ok := true; ok; ok = t.kind != TokenKindEnd {
		instr := p.parseInstr()
		program.instrs = append(program.instrs, instr)

		p.current(&t)
	}

	return program
}

func (p *Parser) parseLabel() NodeInstrLabel {
	var instr NodeInstrLabel

	var t Token
	p.current(&t)
	if t.kind != TokenKindLabel {
		panic(fmt.Errorf("expected label, found: %v", t.kind))
	}
	p.advance()

	instr.name = t.val

	return instr
}

func (p *Parser) parseOutput() NodeInstrOutput {
	var instr NodeInstrOutput

	var t Token
	p.current(&t)
	if t.kind != TokenKindOutput {
		panic(fmt.Errorf("expected output, found: %v", t.kind))
	}
	p.advance()

	term := p.parseTerm()
	instr.term = term

	return instr
}

func (p *Parser) parseGoto() NodeInstrGoto {
	var instr NodeInstrGoto

	var t Token
	p.current(&t)
	if t.kind != TokenKindGoto {
		panic(fmt.Errorf("expected goto, found: %v", t.kind))
	}
	p.advance()

	p.current(&t)
	if t.kind != TokenKindLabel {
		panic(fmt.Errorf("expected label, found: %v", t.kind))
	}
	p.advance()
	instr.label = t.val

	return instr

}

func (p *Parser) parseIf() NodeInstrIf {
	var instr NodeInstrIf

	var t Token
	p.current(&t)
	if t.kind != TokenKindIf {
		panic(fmt.Errorf("expected if, found: %v", t.kind))
	}
	p.advance()

	instr.rel = p.parseRel()

	p.current(&t)
	if t.kind != TokenKindThen {
		panic(fmt.Errorf("expected then, found: %v", t.kind))
	}
	p.advance()

	instr.instr = p.parseInstr()

	return instr
}

func (p *Parser) parseInstr() NodeInstr {
	var instr NodeInstr

	var t Token
	p.current(&t)
	switch {
	case t.kind == TokenKindIdent:
		instr = p.parseAssign()
	case t.kind == TokenKindIf:
		instr = p.parseIf()
	case t.kind == TokenKindGoto:
		instr = p.parseGoto()
	case t.kind == TokenKindOutput:
		instr = p.parseOutput()
	case t.kind == TokenKindLabel:
		instr = p.parseLabel()
	default:
		panic(fmt.Errorf("unexpected token kind: %v", t.kind))
	}

	return instr
}

func (p *Parser) parseRel() NodeRel {
	var rel NodeRel
	var lhs, rhs NodeTerm

	lhs = p.parseTerm()

	var t Token
	p.current(&t)
	switch {
	case t.kind == TokenKindLessThan:
		p.advance()
		rhs = p.parseTerm()
		rel = NodeRelLessThan{lhs: lhs, rhs: rhs}
	case t.kind == TokenKindLessThanEqual:
		p.advance()
		rhs = p.parseTerm()
		rel = NodeRelLessThanEqual{lhs: lhs, rhs: rhs}
	case t.kind == TokenKindDoubleEqual:
		p.advance()
		rhs = p.parseTerm()
		rel = NodeRelEqual{lhs: lhs, rhs: rhs}
	default:
		panic(fmt.Errorf("expected rel token (lessthan, lessthanequal, doubleequal), found: %v", t.kind))
	}

	return rel
}

func (p *Parser) parseAssign() NodeInstrAssign {
	var instr NodeInstrAssign

	var t Token
	p.current(&t)
	if t.kind != TokenKindIdent {
		panic(fmt.Errorf("expected ident, found: %v", t.kind))
	}
	p.advance()
	instr.ident = t.val

	p.current(&t)
	if t.kind != TokenKindEqual {
		panic(fmt.Errorf("expected equal, found: %v", t.kind))
	}
	p.advance()

	instr.expr = p.parseExpr()

	return instr
}

func (p *Parser) parseExpr() NodeExpr {
	var expr NodeExpr

	var lhs, rhs NodeTerm
	lhs = p.parseTerm()

	var t Token
	p.current(&t)
	switch {
	case t.kind == TokenKindPlus:
		p.advance()
		rhs = p.parseTerm()
		expr = NodeExprPlus{lhs: lhs, rhs: rhs}
	case t.kind == TokenKindMinus:
		p.advance()
		rhs = p.parseTerm()
		expr = NodeExprMinus{lhs: lhs, rhs: rhs}
	case t.kind == TokenKindMul:
		p.advance()
		rhs = p.parseTerm()
		expr = NodeExprMul{lhs: lhs, rhs: rhs}
	default:
		expr = NodeExprSingle{lhs}
	}

	return expr
}

func (p *Parser) parseTerm() NodeTerm {
	var term NodeTerm

	var t Token
	p.current(&t)
	switch {
	case t.kind == TokenKindInput:
		term = NodeTermInput{}
	case t.kind == TokenKindInt:
		term = NodeTermInt{val: t.val}
	case t.kind == TokenKindIdent:
		term = NodeTermIdent{val: t.val}
	default:
		panic(fmt.Errorf("expected term token (input, int, or ident), found: %v", t.kind))
	}
	p.advance()

	return term
}
