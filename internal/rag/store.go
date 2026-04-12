package rag

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

type VectorStoreConfig struct {
	ChromaURL      string
	CollectionName string
	Embedder       embeddings.Embedder
}

type VectorStore struct {
	store    chroma.Store
	embedder embeddings.Embedder
}

func NewVectorStore(ctx context.Context, cfg VectorStoreConfig) (*VectorStore, error) {
	chromaURL := cfg.ChromaURL
	if chromaURL == "" {
		chromaURL = os.Getenv("CHROMA_URL")
		if chromaURL == "" {
			chromaURL = "http://localhost:8000"
		}
	}

	collectionName := cfg.CollectionName
	if collectionName == "" {
		collectionName = "ai-dev-cli-db"
	}

	opts := []chroma.Option{
		chroma.WithChromaURL(chromaURL),
		chroma.WithNameSpace(collectionName),
	}

	if cfg.Embedder != nil {
		opts = append(opts, chroma.WithEmbedder(cfg.Embedder))
	}

	store, err := chroma.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create chroma store: %w", err)
	}

	return &VectorStore{
		store:    store,
		embedder: cfg.Embedder,
	}, nil
}

func (vs *VectorStore) AddDocuments(ctx context.Context, docs []schema.Document) error {
	if len(docs) == 0 {
		return nil
	}

	ids, err := vs.store.AddDocuments(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}

	if len(ids) > 0 {
		fmt.Printf("Indexed %d documents\n", len(ids))
	}

	return nil
}

func (vs *VectorStore) Search(ctx context.Context, query string, numResults int) ([]schema.Document, error) {
	if numResults == 0 {
		numResults = 3
	}

	docs, err := vs.store.SimilaritySearch(ctx, query, numResults)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return docs, nil
}

func (vs *VectorStore) SearchWithScore(ctx context.Context, query string, numResults int, scoreThreshold float32) ([]schema.Document, error) {
	if numResults == 0 {
		numResults = 3
	}

	opts := []vectorstores.Option{
		vectorstores.WithScoreThreshold(scoreThreshold),
	}

	docs, err := vs.store.SimilaritySearch(ctx, query, numResults, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return docs, nil
}

func (vs *VectorStore) RemoveCollection(ctx context.Context) error {
	return vs.store.RemoveCollection()
}

func FormatContext(docs []schema.Document) string {
	var ctx strings.Builder
	ctx.WriteString("=== Contexto Recuperado de la Base de Datos Vectorial ===\n\n")

	for i, doc := range docs {
		source := ""
		if src, ok := doc.Metadata["source"].(string); ok {
			source = src
		}
		ctx.WriteString(fmt.Sprintf("[Fragmento %d] (Fuente: %s)\n", i+1, source))
		ctx.WriteString(doc.PageContent)
		ctx.WriteString("\n\n---\n\n")
	}

	return ctx.String()
}
