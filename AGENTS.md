# AGENTS.md - AI Dev CLI Development Guide

This file provides development guidelines for agentic coding agents working on the AI Dev CLI codebase.

## Project Overview

AI Dev CLI is a Go CLI tool that uses AI (OpenAI/Ollama) to improve code and generate tests. Built with Cobra, Viper, and LangChain-go.

**Tech Stack:**
- Go 1.24.4
- Cobra (CLI framework)
- Viper (config)
- LangChain-go (AI integration)
- Testify (testing)

---

## Build, Lint & Test Commands

### Building
```bash
# Build binary to bin/ai-dev-cli
make build

# Or directly
go build -o bin/ai-dev-cli .
```

### Testing
```bash
# Run all tests
make test
go test -v ./...

# Run unit tests only (recommended for fast feedback)
make test-unit
go test -v ./internal/...

# Run integration tests (requires API keys, slower)
make test-integration
SKIP_INTEGRATION=0 go test -v ./tests/integration/...

# Run a single test by name
go test -v -run TestPrompter_ImproveCode ./internal/ai/...

# Run tests matching a pattern
go test -v -run "Test.*" ./internal/...
```

### Linting & Formatting
```bash
# Run linter
make lint
golangci-lint run ./...

# Auto-fix linting issues
make lint-fix
golangci-lint run ./... --fix

# Format code
make fmt
go fmt ./...
gofumpt -w .
```

### Other Commands
```bash
# Clean build artifacts
make clean

# Download and tidy dependencies
make install-deps

# Run the improve command (for testing)
make run-improve ARGS=path/to/file.go

# Run the test generation command
make run-test ARGS=path/to/file.go
```

---

## Code Style Guidelines

### Import Organization

Imports must be grouped and sorted:

```go
import (
    // Standard library
    "context"
    "fmt"
    "os"

    // External packages
    "github.com/spf13/cobra"
    "github.com/tmc/langchaingo/llms"
)
```

Run `go fmt ./...` or use GoLand's Organize Imports to enforce.

### Naming Conventions

- **Files**: snake_case.go (e.g., `client.go`, `prompter_test.go`)
- **Packages**: lowercase, short (e.g., `ai`, `utils`, `config`)
- **Interfaces**: PascalCase with `er` suffix for single-method interfaces (e.g., `AIClient`, `Reader`)
- **Functions/Variables**: camelCase (e.g., `newClient`, `filePath`)
- **Exported Types/Functions**: PascalCase (e.g., `NewClient`, `AIClient`)
- **Constants**: PascalCase or CamelCase depending on export (e.g., `MaxRetries`)
- **Errors**: `Err` prefix for error variables (e.g., `ErrNoProvider`)

### Type Definitions

Place types after imports, before functions:

```go
package ai

import (
    "context"
    "errors"
    "os"

    "github.com/tmc/langchaingo/llms"
)

var ErrNoProvider = errors.New("no AI provider configured")

type AIClient interface {
    Generate(ctx context.Context, prompt string) (string, error)
}

type client struct {
    llm llms.Model
}
```

### Error Handling

- Use `fmt.Errorf` with `%w` for error wrapping:
  ```go
  if err != nil {
      return fmt.Errorf("failed to read file: %w", err)
  }
  ```

- Use `errors.New` for sentinel errors:
  ```go
  var ErrNoProvider = errors.New("no AI provider configured")
  ```

- Return errors early, avoid `else` blocks after errors.

### Function Design

- Keep functions small and focused (< 30 lines where possible)
- Use interfaces for dependencies (e.g., `AIClient` interface in `internal/ai/`)
- Return concrete types when possible, interfaces when flexibility needed
- Add context as first parameter for operations that may be slow:
  ```go
  func (c *client) Generate(ctx context.Context, prompt string) (string, error)
  ```

### Cyclomatic Complexity

Max complexity: 15 (enforced by gocyclo linter)

- Break complex functions into smaller helpers
- Use early returns to reduce nesting

---

## Testing Guidelines

### Test File Naming

Tests must be `*_test.go` in the same package or `tests/` directory.

### Test Structure

Use testify/require:

```go
package ai

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
)

func TestPrompter_ImproveCode(t *testing.T) {
    mockClient := &MockAIClient{
        Response: "improved code",
    }

    prompter := NewPrompter(mockClient)
    result, err := prompter.ImproveCode("func example() {}")

    require.NoError(t, err)
    require.Equal(t, "improved code", result)
}
```

### Mocking

- Define mock types in the same test file or `mock_*.go` files
- Use struct embedding for simpler mocks

```go
type MockAIClient struct {
    Response string
    Err      error
}

func (m *MockAIClient) Generate(ctx context.Context, prompt string) (string, error) {
    if m.Err != nil {
        return "", m.Err
    }
    return m.Response, nil
}
```

### Integration Tests

- Mark with `SKIP_INTEGRATION` environment variable
- Use `t.Skip()` for conditional skipping

```go
func TestCLI_ImproveCommand(t *testing.T) {
    if os.Getenv("SKIP_INTEGRATION") == "1" {
        t.Skip("Skipping integration test")
    }
    // ...
}
```

---

## Linter Configuration

The project uses golangci-lint with these settings (`.golangci.yml`):

**Enabled Linters:**
- errcheck, govet, ineffassign, staticcheck, unused
- gocyclo (max complexity: 15)
- goconst, misspell, revive, unconvert, unparam
- nakedret, bodyclose, errorlint, copyloopvar, noctx
- gosec, gocritic

**Key Rules:**
- `package-comments`: disabled (no requirement for package docs)
- `exported`: enabled (exported functions need docs)
- Stuttering in function names: allowed

Always run `make lint` before committing.

---

## File Organization

```
ai-dev-cli/
├── cmd/           # Cobra commands (root.go, improve.go, test.go)
├── internal/       # Internal packages (ai/, config/)
│   └── ai/        # AI client, prompter logic
├── pkg/           # Reusable packages
├── platform/      # Platform-specific code
├── tests/         # Integration tests
│   └── integration/
├── examples/      # Example files
└── Makefile       # Build commands
```

---

## Environment Variables

Create a `.env` file from `.env.example`:

```bash
AI_PROVIDER=openai  # or "ollama"
OPENAI_API_KEY=your-key-here
OLLAMA_MODEL=llama3.2  # optional, for Ollama
```

---

## Common Development Patterns

### Adding a New Command

1. Create `cmd/<command>.go`:
   ```go
   var <command>Cmd = &cobra.Command{
       Use:   "<command> <file>",
       Short: "Description",
       Args:  cobra.ExactArgs(1),
       RunE: func(cmd *cobra.Command, args []string) error {
           // implementation
           return nil
       },
   }
   ```

2. Register in `cmd/root.go`:
   ```go
   rootCmd.AddCommand(<command>Cmd)
   ```

### Adding a New Package

1. Create directory with meaningful name in `internal/` or `pkg/`
2. Add tests in same directory
3. Update imports using full module path: `github.com/ai-dev-cli/ai-dev-cli/internal/...`

---

## Key Dependencies

- **github.com/spf13/cobra** - CLI commands
- **github.com/spf13/viper** - Configuration
- **github.com/tmc/langchaingo/llms** - AI LLM integration
- **github.com/tmc/langchaingo/vectorstores/chroma** - ChromaDB vector store
- **github.com/tmc/langchaingo/textsplitter** - Text splitting for RAG
- **github.com/stretchr/testify** - Testing assertions

---

## RAG (Retrieval Augmented Generation)

The CLI supports RAG for context-aware code improvements using ChromaDB vector database.

### Quick Start

```bash
# 1. Start ChromaDB
make up

# 2. Index documentation
ai-dev index --path ./docs

# 3. Improve code with RAG context
ai-dev improve --rag main.go
```

### Environment Variables

```bash
# Required for RAG
AI_PROVIDER=openai  # or "ollama"
OPENAI_API_KEY=your-key-here

# Optional
CHROMA_URL=http://localhost:8000  # default
EMBEDDER_MODEL=text-embedding-3-small  # OpenAI default
OLLAMA_BASE_URL=http://localhost:11434  # Ollama default
```

### RAG Commands

```bash
# Index local files or directories
ai-dev index --path ./docs
ai-dev index --path ./internal/ai/

# Index URLs
ai-dev index --url https://docs.example.com/api

# Index with custom collection
ai-dev index --path ./docs --collection my-docs

# Improve with RAG context
ai-dev improve --rag main.go

# Improve with custom provider
ai-dev improve --rag --provider ollama main.go
```

### Code Architecture

```
internal/rag/
├── config.go      # Embedder configuration (OpenAI/Ollama)
├── processor.go   # Document processing (files, URLs, text splitting)
├── store.go       # ChromaDB vector store connection
├── rag.go         # RAGService interface
└── rag_test.go    # Unit tests
```

### Adding RAG to New Commands

```go
// Example: Using RAG in a new command
import (
    "github.com/ai-dev-cli/ai-dev-cli/internal/rag"
)

func runRAGCommand() error {
    cfg := rag.RAGConfig{
        Provider:       "openai",
        CollectionName: "ai-dev-cli-db",
    }
    
    service, err := rag.NewRAGService(ctx, cfg)
    if err != nil {
        return err
    }
    
    // Search for context
    context, err := service.SearchContext(ctx, "your query")
    // Use context in your prompt...
    
    return nil
}
```

---

## Notes for Agents

1. Always run `make lint` and `make test-unit` before finishing any task
2. Keep cyclomatic complexity under 15
3. Write tests for new functionality (unit tests preferred)
4. Use interfaces for testable dependencies
5. Handle errors gracefully with proper wrapping
6. Follow the import grouping convention
7. For single test runs: `go test -v -run TestName ./path/...`