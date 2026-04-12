# Library Documentation

Este documento contiene notas y referencias sobre las librerías externas utilizadas en el proyecto.

---

## Cobra (github.com/spf13/cobra)

**Propósito**: Framework para CLI
**Versión**: v1.9.1

### Conceptos Clave

Cobra se basa en dos conceptos:
- **Commands**: Acciones que el usuario puede ejecutar
- **Flags**: Parámetros de configuración

### Ejemplo de Comando

```go
var improveCmd = &cobra.Command{
    Use:   "improve <file>",
    Short: "Improve code quality using AI",
    Long:  `Analyzes the given Go file...`,
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Lógica del comando
        return nil
    },
}

func init() {
    rootCmd.AddCommand(improveCmd)
}
```

### Flags

| Flag Type | Descripción | Ejemplo |
|-----------|-------------|---------|
| `PersistentFlags` | Disponible en comando y subcomandos | `--config` |
| `LocalFlags` | Solo en el comando específico | `--verbose` |

### Documentación

- https://pkg.go.dev/github.com/spf13/cobra
- https://github.com/spf13/cobra

---

## Viper (github.com/spf13/viper)

**Propósito**: Gestión de configuración
**Versión**: v1.19.0

### Features

- Leer de múltiples fuentes: env vars, flags, archivos
- Hot reload de configuración
- Valores por defecto

### Uso Típico

```go
viper.SetConfigName("config") // sin extensión
viper.AddConfigPath(".")
viper.AutomaticEnv()

if err := viper.ReadInConfig(); err != nil {
    log.Fatalf("Error reading config: %v", err)
}
```

### Variables de Entorno con Prefijo

```go
viper.SetEnvPrefix("APP")  // APP_* becomes viper keys
```

---

## LangChain-go (github.com/tmc/langchaingo)

**Propósito**: Integración con LLMs y RAG
**Versión**: v0.1.14

### Módulos Principales

| Módulo | Descripción |
|--------|-------------|
| `llms` | Comunicación con LLMs (OpenAI, Ollama, Anthropic) |
| `embeddings` | Crear embeddings de texto |
| `vectorstores` | Almacenamiento vectorial (Chroma, Pinecone) |
| `textsplitter` | División de texto en chunks |

### LLMs Soportados

- **OpenAI**: GPT-4, GPT-3.5
- **Ollama**: Modelos locales (llama3.2, etc.)
- **Anthropic**: Claude family
- **Google**: Gemini

### Embeddings

```go
llm, _ := openai.New()
embedder, _ := embeddings.NewEmbedder(llm)
```

### Vector Stores

```go
store, _ := chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),
    chroma.WithEmbedder(embedder),
)
```

---

## ChromaDB (github.com/amikos-tech/chroma-go)

**Propósito**: Vector store para RAG
**Versión**: v0.1.4 (cliente Go)

### Ejecutar Localmente

```bash
docker run -p 8000:8000 chromadb/chroma:latest
```

### API

- **HTTP**: `http://localhost:8000`
- **SDK Go**: `github.com/amikos-tech/chroma-go`

### Colecciones

ChromaDB organiza vectores en colecciones con nombre único:

```go
store, _ := chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),
    chroma.WithNameSpace("mi-proyecto"),
)
```

---

## Testify (github.com/stretchr/testify)

**Propósito**: Testing framework
**Versión**: v1.10.0

### Paquetes

| Paquete | Uso |
|---------|-----|
| `require` | Assertions que paran en el primer error |
| `assert` | Assertions que continue проверка после failure |
| `mock` | Mocking framework |
| `suite` | Test suites |

### Ejemplo

```go
import "github.com/stretchr/testify/require"

func TestExample(t *testing.T) {
    require.NoError(t, err)
    require.Equal(t, expected, actual)
}
```

---

## TextSplitter (github.com/tmc/langchaingo/textsplitter)

**Propósito**: Dividir texto en chunks para embeddings

### RecursiveCharacter

```go
splitter := textsplitter.NewRecursiveCharacter(
    textsplitter.WithChunkSize(1000),
    textsplitter.WithChunkOverlap(200),
)

docs, _ := textsplitter.CreateDocuments(splitter, texts, metadatas)
```

### Opciones

| Opción | Default | Descripción |
|--------|---------|-------------|
| `ChunkSize` | 4000 | Caracteres por chunk |
| `ChunkOverlap` | 200 | Overlap entre chunks |
| `Separators` | ["\n\n", "\n", " ", ""] | Separadores en orden de prioridad |

---

## Go Environment Variables

### AI Providers

| Variable | Descripción | Default |
|----------|-------------|---------|
| `AI_PROVIDER` | "openai" o "ollama" | "openai" |
| `OPENAI_API_KEY` | API key de OpenAI | - |
| `OLLAMA_MODEL` | Modelo de Ollama | "llama3.2" |
| `OLLAMA_BASE_URL` | URL de Ollama | "http://localhost:11434" |
| `CHROMA_URL` | URL de ChromaDB | "http://localhost:8000" |

### Embedding Models

| Provider | Model | Dimensiones |
|----------|-------|-------------|
| OpenAI | text-embedding-3-small | 1536 |
| OpenAI | text-embedding-3-large | 3072 |
| Ollama | nomic-embed-text | 768 |

---

## Notas de Configuración para Desarrollo

### Ollama

```bash
# Ver modelos disponibles
ollama list

# Pull modelo de embeddings
ollama pull nomic-embed-text
```

### ChromaDB

```bash
# Ver colecciones
curl http://localhost:8000/api/v1/collections
```

---

## Actualización de Dependencias

```bash
# Actualizar todas las dependencias
go get -u ./...

# Actualizar una dependencia específica
go get -u github.com/spf13/cobra

# Limpiar no-used
go mod tidy
```