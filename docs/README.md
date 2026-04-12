# AI Dev CLI - Documentación del Proyecto

## Descripción

AI Dev CLI es una herramienta de línea de comandos que utiliza inteligencia artificial para mejorar código y generar tests automáticamente. Construida con Go, integra LangChain-go para la comunicación con modelos de OpenAI y Ollama.

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
```

## Configuración

Ver archivo `.env.example` para las variables de entorno requeridas:

- `AI_PROVIDER`: openai u ollama
- `OPENAI_API_KEY`: Tu API key de OpenAI
- `CHROMA_URL`: URL de ChromaDB (default: http://localhost:8000)

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