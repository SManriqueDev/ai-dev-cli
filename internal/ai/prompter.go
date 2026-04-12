package ai

import (
	"context"
	"strings"
	"text/template"
)

type Prompter struct {
	client AIClient
}

func NewPrompter(client AIClient) *Prompter {
	return &Prompter{client: client}
}

func (p *Prompter) ImproveCode(code string) (string, error) {
	prompt := buildImprovePrompt(code, "")
	return p.client.Generate(context.Background(), prompt)
}

func (p *Prompter) ImproveCodeWithContext(code, ragContext string) (string, error) {
	prompt := buildImprovePrompt(code, ragContext)
	return p.client.Generate(context.Background(), prompt)
}

func (p *Prompter) GenerateTests(code string) (string, error) {
	prompt := buildTestPrompt(code, "")
	return p.client.Generate(context.Background(), prompt)
}

// nolint:misspell
var improveTemplate = template.Must(template.New("improve").Parse(`
You are a Go senior developer and code reviewer. Analyze the following code and provide
constructive improvements following Chain-of-Thought reasoning.

{{if .Context}}
=== Contexto Recuperado de la Base de Datos Vectorial ===
El siguiente contexto ha sido recuperado de la documentación indexada. Úsalo como referencia
para entender las convenciones y patrones usados en este proyecto:

{{.Context}}
=== Fin del Contexto ===
{{end}}

Analyze step by step:
1. Identify code smells and anti-patterns
2. Suggest performance optimizations
3. Recommend best practices improvements
4. Point out potential bugs or edge cases

Original code:
{{.Code}}

Provide the improved version with explanations for each change.
`))

var testTemplate = template.Must(template.New("test").Parse(`
You are a Go senior developer. Generate comprehensive unit tests for the following code
using few-shot prompting with examples.

{{if .Context}}
=== Contexto Recuperado de la Base de Datos Vectorial ===
{{.Context}}
=== Fin del Contexto ===
{{end}}

Requirements:
1. Use testify/require for assertions
2. Include table-driven tests where appropriate
3. Test edge cases and error conditions
4. Follow Go testing conventions

Original code:
{{.Code}}

Generate the test file with proper imports and package naming (_test suffix).
`))

func buildImprovePrompt(code, context string) string {
	var sb strings.Builder
	_ = improveTemplate.Execute(&sb, struct {
		Code    string
		Context string
	}{Code: code, Context: context})
	return sb.String()
}

func buildTestPrompt(code string, context string) string {
	var sb strings.Builder
	_ = testTemplate.Execute(&sb, struct {
		Code    string
		Context string
	}{Code: code, Context: context})
	return sb.String()
}
