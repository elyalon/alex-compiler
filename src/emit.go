package main

import (
	"fmt"
	"os"
	"slices"
)

type Scope struct {
	idents []string
}

func createScope() Scope {
	var scope Scope
	scope.idents = make([]string, 0)
	return scope
}

func (scope *Scope) append(ident string) {
	scope.idents = append(scope.idents, ident)
}

func (scope *Scope) has(ident string) bool {
	return slices.Contains(scope.idents, ident)
}

func (scope *Scope) find(queryIdent string) int {
	for i, ident := range scope.idents {
		if ident == queryIdent {
			return i
		}
	}
	panic(fmt.Errorf("cannot find ident in scope: %v", queryIdent))
}

func emitProgram(program *NodeProgram) {
	scope := createScope()
	for _, instr := range program.instrs {
		scopeInstr(instr, &scope)
	}
	fmt.Fprintf(os.Stderr, "Scope: %v\n", scope.idents)

	fmt.Printf("format ELF64 executable\n")

	// Constants
	fmt.Printf("LINE_MAX equ 1024\n")

	fmt.Printf("segment readable executable\n")
	fmt.Printf("include \"src/lib.asm\"\n")
	fmt.Printf("entry _start\n")
	fmt.Printf("_start:\n")

	fmt.Printf("    mov rbp, rsp\n")
	fmt.Printf("    sub rsp, %d\n", len(scope.idents)*8)

	ifCount := 0
	for _, instr := range program.instrs {
		emitInstr(instr, &scope, &ifCount)
	}

	fmt.Printf("    add rsp, %d\n", len(scope.idents)*8)

	// exit with code 0
	fmt.Printf("    mov rax, 60\n")
	fmt.Printf("    mov rdi, 0\n")
	fmt.Printf("    syscall\n")

	// Data
	fmt.Printf("segment readable writeable\n")
	fmt.Printf("newline db 0xa\n")
	fmt.Printf("line rb LINE_MAX\n")
}

func emitInstr(instr NodeInstr, scope *Scope, ifCount *int) {
	switch instr := instr.(type) {
	case NodeInstrAssign:
		emitExpr(instr.expr, scope)
		index := scope.find(instr.ident)
		fmt.Printf("    mov qword [rbp - %d], rax ; Store in `%v`\n", index*8+8, instr.ident)
	case NodeInstrIf:
		emitRel(instr.rel, scope)
		suf := *ifCount
		(*ifCount)++
		fmt.Printf("    test rax, rax\n")
		fmt.Printf("    jz .endif%d\n", suf)
		emitInstr(instr.instr, scope, ifCount)
		fmt.Printf(".endif%d:\n", suf)
	case NodeInstrGoto:
		fmt.Printf("    jmp .%s\n", instr.label)
	case NodeInstrOutput:
		emitTerm(instr.term, scope)
		fmt.Printf("    mov rdi, 1\n") // stdout
		fmt.Printf("    mov rsi, rax\n")
		fmt.Printf("    call write_uint\n")
		fmt.Printf("    write 1, newline, 1\n")
	case NodeInstrLabel:
		fmt.Printf(".%s:\n", instr.name)
	}
}

func emitRel(rel NodeRel, scope *Scope) {
	switch rel := rel.(type) {
	case NodeRelLessThan:
		emitTerm(rel.lhs, scope)
		fmt.Printf("    mov r12, rax\n")
		emitTerm(rel.rhs, scope)
		fmt.Printf("    cmp r12, rax\n")
		fmt.Printf("    setl al\n")
		fmt.Printf("    and al, 1\n")
		fmt.Printf("    movzx rax, al\n")
	case NodeRelLessThanEqual:
		emitTerm(rel.lhs, scope)
		fmt.Printf("    mov r12, rax\n")
		emitTerm(rel.rhs, scope)
		fmt.Printf("    cmp r12, rax\n")
		fmt.Printf("    setle al\n")
		fmt.Printf("    and al, 1\n")
		fmt.Printf("    movzx rax, al\n")
	case NodeRelEqual:
		emitTerm(rel.lhs, scope)
		fmt.Printf("    mov r12, rax\n")
		emitTerm(rel.rhs, scope)
		fmt.Printf("    cmp r12, rax\n")
		fmt.Printf("    sete al\n")
		fmt.Printf("    and al, 1\n")
		fmt.Printf("    movzx rax, al\n")
	}
}

func emitExpr(expr NodeExpr, scope *Scope) {
	switch expr := expr.(type) {
	case NodeExprSingle:
		emitTerm(expr.term, scope)
	case NodeExprPlus:
		emitTerm(expr.lhs, scope)
		fmt.Printf("    mov r12, rax\n")
		emitTerm(expr.rhs, scope)
		fmt.Printf("    add r12, rax\n")
		fmt.Printf("    mov rax, r12\n")
	case NodeExprMinus:
		emitTerm(expr.lhs, scope)
		fmt.Printf("    mov r12, rax\n")
		emitTerm(expr.rhs, scope)
		fmt.Printf("    sub r12, rax\n")
		fmt.Printf("    mov rax, r12\n")
	case NodeExprMul:
		emitTerm(expr.lhs, scope)
		fmt.Printf("    mov r12, rax\n")
		emitTerm(expr.rhs, scope)
		fmt.Printf("    imul r12, rax\n")
		fmt.Printf("    mov rax, r12\n")
	}
}

func emitTerm(term NodeTerm, scope *Scope) {
	switch term := term.(type) {
	case NodeTermInput:
		fmt.Printf("    read 0, line, LINE_MAX\n")
		fmt.Printf("    mov rdi, line\n")
		fmt.Printf("    call strlen\n")
		fmt.Printf("    mov rdi, line\n")
		fmt.Printf("    mov rsi, rax\n")
		fmt.Printf("    call parse_uint\n")
	case NodeTermInt:
		fmt.Printf("    mov rax, %s\n", term.val)
	case NodeTermIdent:
		index := scope.find(term.val)
		fmt.Printf("    mov rax, qword [rbp - %d] ; Load `%v`\n", index*8+8, term.val)
	}
}

// `scopeNode` functions check the scope of every ident in that node
// -- `scopeInstr` also appends idents to the scope

func scopeInstr(instr NodeInstr, scope *Scope) {
	switch instr := instr.(type) {
	case NodeInstrAssign:
		scopeExpr(instr.expr, scope)
		if !scope.has(instr.ident) {
			scope.append(instr.ident)
		}
	case NodeInstrIf:
		scopeRel(instr.rel, scope)
		scopeInstr(instr.instr, scope)
	case NodeInstrGoto:
	case NodeInstrOutput:
		scopeTerm(instr.term, scope)
	case NodeInstrLabel:
	}
}

func scopeRel(rel NodeRel, scope *Scope) {
	switch rel := rel.(type) {
	case NodeRelLessThan:
		scopeTerm(rel.lhs, scope)
		scopeTerm(rel.rhs, scope)
	case NodeRelLessThanEqual:
		scopeTerm(rel.lhs, scope)
		scopeTerm(rel.rhs, scope)
	case NodeRelEqual:
		scopeTerm(rel.lhs, scope)
		scopeTerm(rel.rhs, scope)
	}
}

func scopeExpr(expr NodeExpr, scope *Scope) {
	switch expr := expr.(type) {
	case NodeExprSingle:
		scopeTerm(expr.term, scope)
	case NodeExprPlus:
		scopeTerm(expr.lhs, scope)
		scopeTerm(expr.rhs, scope)
	case NodeExprMinus:
		scopeTerm(expr.lhs, scope)
		scopeTerm(expr.rhs, scope)
	case NodeExprMul:
		scopeTerm(expr.lhs, scope)
		scopeTerm(expr.rhs, scope)
	}
}

func scopeTerm(term NodeTerm, scope *Scope) {
	switch term := term.(type) {
	case NodeTermInput:
	case NodeTermInt:
	case NodeTermIdent:
		if !scope.has(term.val) {
			panic(fmt.Errorf("ident not defined: %v", term.val))
		}
	}
}
