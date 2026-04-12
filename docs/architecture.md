# Architecture Decision Records (ADR)

Este documento describe las decisiones arquitectónicas tomadas en el proyecto AI Dev CLI.

## ADR-001: Arquitectura de Capas

**Estado**: Aprobado
**Fecha**: 2024-01-15

### Contexto

Necesitamos una estructura de proyecto que sea mantenible, testeable y que permita escalar el CLI con nuevos comandos y funcionalidades de IA.

### Decisión

Usaremos una arquitectura de capas inspirada en Clean Architecture:

```
ai-dev-cli/
├── cmd/              # Punto de entrada - comandos Cobra
├── internal/         # Paquetes internos (no exportables)
│   ├── ai/          # Lógica de IA (client, prompter)
│   └── rag/         # RAG y vector store
├── pkg/             # Paquetes reutilizables (exportables)
├── platform/        # Código platform-specific
├── tests/           # Tests de integración
└── examples/        # Archivos de ejemplo
```

### Consecuencias

**Positivas**:
- Separación clara de responsabilidades
- Facilita testing unitario
- Permite swapping de implementaciones (ej: diferentes AI providers)

**Negativas**:
- Más boilerplate inicial
- Curva de aprendizaje para nuevos contribuidores

---

## ADR-002: Stack de AI - LangChain-go

**Estado**: Aprobado
**Fecha**: 2024-01-15

### Contexto

Necesitamos una biblioteca Go que abstraiga la comunicación con diferentes LLMs (OpenAI, Ollama, Anthropic, etc.).

### Decisión

Usar **LangChain-go** (github.com/tmc/langchaingo) como biblioteca principal para:
- Comunicación con LLMs
- Embeddings para RAG
- Vector stores (ChromaDB)

### Alternativas Consideradas

| Alternativa | Pros | Contras |
|-------------|------|---------|
| go-openai | Simple, liviano | Solo OpenAI |
| Ollama-go | Solo Ollama | Limitado |
| LangChain-go | Multi-provider, RAG | Complejo |

### Consecuencias

**Positivas**:
- Soporte para múltiples providers
- Integración nativa con ChromaDB
- Comunidad activa

**Negativas**:
- API puede cambiar entre versiones
- Algunas features tienen wrapper inestable

---

## ADR-003: Vector Store - ChromaDB

**Estado**: Aprobado
**Fecha**: 2024-01-15

### Contexto

Para RAG necesitamos una base de datos vectorial que corra localmente y sea liviana.

### Decisión

Usar **ChromaDB** (via Docker) como vector store:
- Corre localmente en Docker
- API REST simple
- Integración nativa con LangChain-go

### Justificación

| Opción | Pros | Contras |
|--------|------|---------|
| ChromaDB | Local, liviano, LangChain | Requiere Docker |
| Pinecone | Cloud, robusto | Requiere cuenta |
| Qdrant | Local option | Menos integrado |
| Weaviate | Flexible | Más complejo |

ChromaDB es la mejor opción para un CLI que corre localmente.

---

## ADR-004: Patrón de Templates para Prompts

**Estado**: Aprobado
**Fecha**: 2024-01-15

### Contexto

Los prompts para mejorar código y generar tests necesitan ser maintainables y permitir configuración.

### Decisión

Usar `text/template` de Go para construir prompts:
- Permite variables dinámicas
- Separación entre lógica y contenido
- Testing de templates posible

### Ejemplo

```go
var improveTemplate = template.Must(template.New("improve").Parse(`
You are a Go senior developer. Analyze the following code:
{{.Code}}
{{if .Context}}
Context: {{.Context}}
{{end}}
`))
```

---

## ADR-005: Testing con Testify

**Estado**: Aprobado
**Fecha**: 2024-01-15

### Contexto

Necesitamos un framework de testing que facilite mocking y assertions claras.

### Decisión

Usar **testify** (github.com/stretchr/testify):
- `require` para fail-fast
- `assert` para verificaciones adicionales
- Mocking simple con struct embedding

### Política de Testing

1. Tests unitarios en el mismo paquete que el código
2. Tests de integración en `tests/integration/`
3. Skip tests que requieren API keys con `SKIP_INTEGRATION`

---

## ADR-006: Flags de Configuración

**Estado**: Aprobado
**Fecha**: 2024-01-15

### Contexto

El CLI debe permitir configuración flexible sin硬codificar valores.

### Decisión

Usar **Viper** para configuración:
- Lee de `.env` por defecto
- Usa `platform/config` como fuente central de defaults y valores activos
- Soporta flags y variables de entorno explícitamente vinculadas
- Configuración hiérarchica

### Orden de Precedencia

1. Flags de línea de comandos
2. Variables de entorno
3. Archivo de configuración
4. Valores por defecto

### Configuración por provider

- `AI_PROVIDER` selecciona el bloque activo (`openai` u `ollama`)
- OpenAI usa `OPENAI_API_KEY`, `OPENAI_BASE_URL`, `OPENAI_MODEL` y `OPENAI_EMBEDDER_MODEL`
- Ollama usa `OLLAMA_BASE_URL`, `OLLAMA_MODEL` y `OLLAMA_EMBEDDER_MODEL`
- Chroma y RAG usan `CHROMA_URL`, `CHROMA_COLLECTION_NAME`, `RAG_CHUNK_SIZE` y `RAG_CHUNK_OVERLAP`
- El proyecto no usa un `EMBEDDER_MODEL` genérico
