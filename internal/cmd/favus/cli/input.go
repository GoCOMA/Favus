package cli

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "strings"
)

func PromptInput(prompt string) string {
    return PromptInputWithIO(prompt, os.Stdin, os.Stdout)
}

func PromptInputWithIO(prompt string, r io.Reader, w io.Writer) string {
    reader := bufio.NewReader(r)
    fmt.Fprint(w, prompt)
    in, _ := reader.ReadString('\n')
    return strings.TrimSpace(in)
}

func PromptRequired(label string) string {
    for {
        v := PromptInput(label) // 기존과 동일하게 label만 넘김
        v = strings.TrimSpace(v)
        if v != "" {
            return v
        }
        fmt.Println("값이 비어있습니다. 다시 입력해주세요.")
    }
}

func PromptWithDefault(label, def string) string {
    v := PromptInput(label + fmt.Sprintf(" (default: %s): ", def))
    v = strings.TrimSpace(v)
    if v == "" {
        return def
    }
    return v
}