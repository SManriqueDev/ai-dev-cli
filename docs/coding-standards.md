# Coding Standards

Este documento describe las convenciones y estándares de código seguidos en el proyecto AI Dev CLI.

## Convenciones de Nomenclatura

### Archivos
- **snake_case.go**: `client.go`, `prompter_test.go`
- Avoid: camelCase, PascalCase

### Paquetes
- **lowercase, short**: `ai`, `rag`, `utils`
- Avoid: nombres largos o descriptivos

### Funciones y Variables
- **camelCase**: `newClient`, `filePath`
- **Exported (exportables)**: PascalCase: `NewClient`, `AIClient`

### Interfaces
- **Single method**: suffix `er`: `AIClient`, `Reader`, `Writer`
- **Multi method**: PascalCase descriptivo: `VectorStore`

### Constantes
- PascalCase si son exportadas: `MaxRetries`
- camelCase si son privadas: `defaultTimeout`

### Errores
- Prefix `Err`: `ErrNoProvider`, `ErrInvalidConfig`

---

## Estructura de Archivos Go

### Order de Declaraciones

```go
package name

import (
    // 1. Standard library
    "context"
    "fmt"
    "os"
    
    // 2. External packages
    "github.com/spf13/cobra"
    "github.com/tmc/langchaingo/llms"
    
    // 3. Internal packages
    "github.com/ai-dev-cli/ai-dev-cli/internal/ai"
)

// Constantes (si hay)
// Variables (si hay)
// Types e Interfaces
// Funciones
```

---

## Patrones de Código

### Error Handling

✅ **Correcto**:
```go
if err != nil {
    return fmt.Errorf("failed to read file: %w", err)
}
```

❌ **Evitar**:
```go
if err != nil {
    fmt.Println(err)
    return nil  // Silent fail
}
```

### Sentinel Errors

```go
var ErrNoProvider = errors.New("no AI provider configured")

func example() error {
    if provider == "" {
        return ErrNoProvider
    }
    // ...
}
```

### Context como Primer Parametro

```go
func (c *client) Generate(ctx context.Context, prompt string) (string, error) {
    // ...
}
```

### Retorno Temprano (Early Returns)

✅ **Correcto**:
```go
func process(data string) error {
    if data == "" {
        return ErrEmptyData
    }
    
    if len(data) > maxLength {
        return ErrTooLong
    }
    
    // Resto de la lógica
    return nil
}
```

❌ **Evitar**:
```go
func process(data string) error {
    if data != "" {
        if len(data) <= maxLength {
            // Lógica aquí
            return nil
        } else {
            return ErrTooLong
        }
    } else {
        return ErrEmptyData
    }
}
```

---

## Testing Standards

### Estructura de Tests

```go
package ai

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestPrompter_ImproveCode(t *testing.T) {
    // Arrange
    mockClient := &MockAIClient{
        Response: "improved code",
    }
    prompter := NewPrompter(mockClient)
    
    // Act
    result, err := prompter.ImproveCode("func example() {}")
    
    // Assert
    require.NoError(t, err)
    require.Equal(t, "improved code", result)
}
```

### Mocking con Struct Embedding

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

### Nombrado de Tests

- `Test<Struct>_<Method>`: `TestPrompter_ImproveCode`
- `Test<Package>_<Description>`: `TestRAG_SearchContext`

---

## Complejidad Ciclomática

**Máximo**: 15 (enforzado por linter gocyclo)

### Estrategias de Reducción

1. **Extraer funciones helpers**:
```go
// Antes
func complex() error {
    if cond1 {
        if cond2 {
            if cond3 {
                // deep nesting
            }
        }
    }
}

// Después
func complex() error {
    if !cond1 {
        return validateCond1()
    }
    if !cond2 {
        return validateCond2()
    }
    return process()
}
```

2. **Usar switch en lugar de if-else largos**:
```go
switch provider {
case "openai":
    return newOpenAI()
case "ollama":
    return newOllama()
default:
    return newOpenAI()
}
```

### Configuración

- Evitar `os.Getenv` en consumidores; preferir `platform/config` como fuente única.
- Cuando una opción dependa de `AI_PROVIDER`, usar variables específicas por provider (`OPENAI_*` / `OLLAMA_*`).
- No introducir variables genéricas nuevas si ya existe un bloque de configuración por provider.

---

## Import Organization

Ejecutar `go fmt ./...` regularmente o usar "Organize Imports" de tu IDE.

### Grupo de Imports

```go
import (
    // Standard library
    "context"
    "fmt"
    "os"
    
    // External packages (orden alfabético)
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/stretchr/testify"
    "github.com/tmc/langchaingo/llms"
    
    // Internal packages
    "github.com/ai-dev-cli/ai-dev-cli/internal/ai"
    "github.com/ai-dev-cli/ai-dev-cli/internal/rag"
)
```

---

## Documentación de Código

### Funciones Exportadas

Todas las funciones exportadas deben tener doc comments:

```go
// NewClient creates a new AI client based on the active provider configuration.
// Returns an error if the provider or client initialization fails.
func NewClient() (AIClient, error) {
    // ...
}
```

### Paquetes

No se requiere doc comment para paquetes, pero es recomendado.

---

## Git Commit Messages

Usar Conventional Commits:

```
feat: add RAG support for context-aware improvements
fix: resolve nil pointer in prompter
docs: update architecture.md with new ADR
test: add unit tests for document processor
refactor: simplify error handling in client.go
```

### Prefijos

- `feat`: Nueva funcionalidad
- `fix`: Bug fix
- `docs`: Documentación
- `test`: Tests
- `refactor`: Refactoring
- `chore`: Mantenimiento
