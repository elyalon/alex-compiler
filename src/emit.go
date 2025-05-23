package main

import (
	"fmt"
	"slices"
)

func (program *NodeProgram) emit() {
	variables := make([]string, 0)
	ifCount := 0

	for _, instr := range program.instrs {
		instrDeclareVariables(instr, &variables)
	}

	fmt.Printf("format ELF64 executable\n")

	// Constants
	fmt.Printf("LINE_MAX equ 1024\n")

	fmt.Printf("segment readable executable\n")
	fmt.Printf("include \"thirdparty/linux.inc\"\n")
	fmt.Printf("include \"thirdparty/utils.inc\"\n")
	fmt.Printf("entry _start\n")
	fmt.Printf("_start:\n")

	fmt.Printf("    mov rbp, rsp\n")
	fmt.Printf("    sub rsp, %d\n", len(variables)*8)

	for _, instr := range program.instrs {
		emitInstr(instr, &variables, &ifCount)
	}

	fmt.Printf("    add rsp, %d\n", len(variables)*8)

	fmt.Printf("    mov rax, 60\n")
	fmt.Printf("    xor rdi, rdi\n")
	fmt.Printf("    syscall\n")

	fmt.Printf("segment readable writeable\n")
	fmt.Printf("newline db 0xa\n")
	fmt.Printf("line rb LINE_MAX\n")
}

func emitInstr(instr NodeInstr, variables *[]string, ifCount *int) {
	switch instr := instr.(type) {
	case NodeInstrAssign:
		emitExpr(instr.expr, variables)
		index := findVariable(variables, instr.ident)
		fmt.Printf("    mov qword [rbp - %d], rax\n", index*8+8)
	case NodeInstrIf:
		emitRel(instr.rel, variables)
		suf := *ifCount
		(*ifCount)++
		fmt.Printf("    test rax, rax\n")
		fmt.Printf("    jz .endif%d\n", suf)
		emitInstr(instr.instr, variables, ifCount)
		fmt.Printf(".endif%d:\n", suf)
	case NodeInstrGoto:
		fmt.Printf("    jmp .%s\n", instr.label)
	case NodeInstrOutput:
		emitTerm(instr.term, variables)
		fmt.Printf("    mov rdi, 1\n") // stdout
		fmt.Printf("    mov rsi, rax\n")
		fmt.Printf("    call write_uint\n")
		fmt.Printf("    write 1, newline, 1\n")
	case NodeInstrLabel:
		fmt.Printf(".%s:\n", instr.name)
	}
}

func emitRel(rel NodeRel, variables *[]string) {
	switch rel := rel.(type) {
	case NodeRelLessThan:
		emitTerm(rel.lhs, variables)
		fmt.Printf("    mov rdx, rax\n")
		emitTerm(rel.rhs, variables)
		fmt.Printf("    cmp rdx, rax\n")
		fmt.Printf("    setl al\n")
		fmt.Printf("    and al, 1\n")
		fmt.Printf("    movzx rax, al\n")
	}
}

func emitExpr(expr NodeExpr, variables *[]string) {
	switch expr := expr.(type) {
	case NodeExprSingle:
		emitTerm(expr.term, variables)
	case NodeExprPlus:
		emitTerm(expr.lhs, variables)
		fmt.Printf("    mov rdx, rax\n")
		emitTerm(expr.rhs, variables)
		fmt.Printf("    add rax, rdx\n")
	}
}

func emitTerm(term NodeTerm, variables *[]string) {
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
		index := findVariable(variables, term.val)
		fmt.Printf("    mov rax, qword [rbp - %d]\n", index*8+8)
	}
}

func findVariable(variables *[]string, ident string) int {
	for i, variable := range *variables {
		if variable == ident {
			return i
		}
	}
	return -1
}

func instrDeclareVariables(instr NodeInstr, variables *[]string) {
	switch instr := instr.(type) {
	case NodeInstrAssign:
		exprDeclareVariables(instr.expr, variables)
		if slices.Contains(*variables, instr.ident) {
			return
		}
		*variables = append(*variables, instr.ident)
	case NodeInstrIf:
		relDeclareVariables(instr.rel, variables)
		instrDeclareVariables(instr.instr, variables)
	case NodeInstrGoto:
	case NodeInstrOutput:
		termDeclareVariables(instr.term, variables)
	case NodeInstrLabel:
	}
}

func relDeclareVariables(rel NodeRel, variables *[]string) {
	switch rel := rel.(type) {
	case NodeRelLessThan:
		termDeclareVariables(rel.lhs, variables)
		termDeclareVariables(rel.rhs, variables)
	}
}

func exprDeclareVariables(expr NodeExpr, variables *[]string) {
	switch expr := expr.(type) {
	case NodeExprSingle:
		termDeclareVariables(expr.term, variables)
	case NodeExprPlus:
		termDeclareVariables(expr.lhs, variables)
		termDeclareVariables(expr.rhs, variables)
	}
}

func termDeclareVariables(term NodeTerm, variables *[]string) {
	switch term := term.(type) {
	case NodeTermInput:
	case NodeTermInt:
	case NodeTermIdent:
		for _, v := range *variables {
			if term.val == v {
				return
			}
		}
		panic(fmt.Errorf("ident not defined: %v", term.val))
	}
}
