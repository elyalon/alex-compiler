macro syscall3 number, a, b, c
{
    mov rax, number
    mov rdi, a
    mov rsi, b
    mov rdx, c
    syscall
}

macro write fd, buf, count
{
    syscall3 1, fd, buf, count
}

macro read fd, buf, count
{
    syscall3 0, fd, buf, count
}

;; Write an integer to a file
;;   rdi - int fd
;;   rsi - uint64_t int x
write_uint:
    test rsi, rsi
    jz .base_zero

    mov rcx, 10     ;; 10 literal for division
    mov rax, rsi    ;; keeping track of rsi in rax cause it's easier to div it like that
    mov r10, 0      ;; counter of how many digits we already converted
.next_digit:
    test rax, rax
    jz .done
    mov rdx, 0
    div rcx
    add rdx, '0'
    dec rsp
    mov byte [rsp], dl
    inc r10
    jmp .next_digit
.done:
    write rdi, rsp, r10
    add rsp, r10
    ret
.base_zero:
    dec rsp
    mov byte [rsp], '0'
    write rdi, rsp, 1
    inc rsp
    ret

;; Compute the length of a NULL-terminated string
;;   rdi - const char *s
strlen:
    push rdi
    xor rax, rax
.next_char:
    mov al, byte [rdi]
    cmp rax, 0
    je .done

    inc rdi
    jmp .next_char
.done:
    pop rsi
    sub rdi, rsi
    mov rax, rdi
    ret

;; Parse unsigned integer from a sized string
;;   rdi - void *buf
;;   rsi - size_t n
parse_uint:
    xor rax, rax
    xor rbx, rbx
    mov rcx, 10
.next_digit:
    cmp rsi, 0
    jle .done

    mov bl, byte [rdi]
    cmp rbx, '0'
    jl .done
    cmp rbx, '9'
    jg .done
    sub rbx, '0'

    mul rcx
    add rax, rbx

    inc rdi
    dec rsi
    jmp .next_digit
.done:
    ret