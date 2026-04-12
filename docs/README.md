# AI Dev CLI - Documentación del Proyecto

## Descripción

AI Dev CLI es una herramienta de línea de comandos que utiliza inteligencia artificial para mejorar código, generar tests automáticamente e indexar documentación para RAG. Construida con Go, integra LangChain-go para la comunicación con modelos de OpenAI y Ollama.

## Características Principales

- **Mejora de Código**: Análisis inteligente de código fuente con sugerencias de mejora
- **Generación de Tests**: Creación automática de tests unitarios usando few-shot prompting
- **RAG (Retrieval Augmented Generation)**: Soporte para mejorar código usando contexto de documentación indexada

## Instalación

```bash
# Clonar el repositorio
git clone https://github.com/ai-dev-cli/ai-dev-cli.git
cd ai-dev-cli

# Instalar dependencias
make install-deps

# Construir el binary
make build
```

## Uso Básico

```bash
# Mejorar código
./bin/ai-dev-cli improve path/to/file.go

# Generar tests
./bin/ai-dev-cli test path/to/file.go

# Indexar documentación para RAG
./bin/ai-dev-cli index --path ./docs

# Mejorar con contexto RAG
./bin/ai-dev-cli improve --rag path/to/file.go

# Forzar provider para RAG o indexación
./bin/ai-dev-cli improve --rag --provider ollama path/to/file.go
./bin/ai-dev-cli index --provider openai --path ./docs
```

## Configuración

La configuración vive en `platform/config` y se resuelve desde `.env`/variables de entorno con defaults explícitos.

### Selección de provider

- `AI_PROVIDER=openai` usa `OPENAI_*`
- `AI_PROVIDER=ollama` usa `OLLAMA_*`
- Los flags `--provider` y `--collection` solo sobreescriben la configuración dentro del comando cuando se pasan explícitamente

### Variables principales

Ver también `.env.example` para la lista completa:

| Variable | Descripción | Default |
| --- | --- | --- |
| `AI_PROVIDER` | Provider activo (`openai` u `ollama`) | `openai` |
| `OPENAI_API_KEY` | API key de OpenAI | - |
| `OPENAI_BASE_URL` | Base URL de OpenAI | `https://api.openai.com/v1` |
| `OPENAI_MODEL` | Modelo de OpenAI para chat | `gpt-4o` |
| `OPENAI_EMBEDDER_MODEL` | Modelo de embeddings para OpenAI | `text-embedding-3-small` |
| `OLLAMA_BASE_URL` | Base URL de Ollama | `http://localhost:11434` |
| `OLLAMA_MODEL` | Modelo de Ollama para chat | `llama3.2` |
| `OLLAMA_EMBEDDER_MODEL` | Modelo de embeddings para Ollama | `nomic-embed-text` |
| `CHROMA_URL` | URL de ChromaDB | `http://localhost:8000` |
| `CHROMA_COLLECTION_NAME` | Nombre de la colección vectorial | `ai-dev-cli-db` |
| `RAG_CHUNK_SIZE` | Tamaño de chunk para RAG | `1000` |
| `RAG_CHUNK_OVERLAP` | Solapamiento entre chunks | `200` |

## Comandos Disponibles

| Comando | Descripción |
|---------|-------------|
| `improve` | Mejora el código dado |
| `test` | Genera tests para el código dado |
| `index` | Indexa documentación en ChromaDB |

## Desarrollo

```bash
# Ejecutar tests
make test

# Ejecutar linter
make lint

# Iniciar ChromaDB
make up
```
