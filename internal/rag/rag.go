package rag

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

type RAGConfig struct {
	ChromaURL      string
	CollectionName string
	Provider       string
	EmbedderModel  string
	APIKey         string
	BaseURL        string
	OllamaURL      string
	ChunkSize      int
	ChunkOverlap   int
	NumResults     int
}

type RAGService interface {
	IndexPath(ctx context.Context, path string) error
	IndexURL(ctx context.Context, url string) error
	IndexText(ctx context.Context, text string, metadata map[string]any) error
	Search(ctx context.Context, query string) ([]schema.Document, error)
	SearchContext(ctx context.Context, query string) (string, error)
}

type ragService struct {
	processor *DocumentProcessor
	store     *VectorStore
}

func NewRAGService(ctx context.Context, cfg RAGConfig) (RAGService, error) {
	chunkSize := cfg.ChunkSize
	if chunkSize == 0 {
		chunkSize = 1000
	}

	chunkOverlap := cfg.ChunkOverlap
	if chunkOverlap == 0 {
		chunkOverlap = 200
	}

	processor := NewDocumentProcessor(chunkSize, chunkOverlap)

	embedder, err := NewEmbedder(ctx, EmbedderConfig{
		Provider:  cfg.Provider,
		Model:     cfg.EmbedderModel,
		APIKey:    cfg.APIKey,
		BaseURL:   cfg.BaseURL,
		OllamaURL: cfg.OllamaURL,
	})
	if err != nil {
		return nil, err
	}

	store, err := NewVectorStore(ctx, VectorStoreConfig{
		ChromaURL:      cfg.ChromaURL,
		CollectionName: cfg.CollectionName,
		Embedder:       embedder,
	})
	if err != nil {
		return nil, err
	}

	return &ragService{
		processor: processor,
		store:     store,
	}, nil
}

func (r *ragService) IndexPath(ctx context.Context, path string) error {
	var docs []schema.Document
	var err error

	if IsValidURL(path) {
		docs, err = r.processor.ProcessURL(path)
	} else {
		docs, err = r.processor.ProcessDirectory(path)
		if err != nil {
			docs, err = r.processor.ProcessFile(path)
		}
	}

	if err != nil {
		return err
	}

	return r.store.AddDocuments(ctx, docs)
}

func (r *ragService) IndexURL(ctx context.Context, url string) error {
	docs, err := r.processor.ProcessURL(url)
	if err != nil {
		return err
	}

	return r.store.AddDocuments(ctx, docs)
}

func (r *ragService) IndexText(ctx context.Context, text string, metadata map[string]any) error {
	docs, err := r.processor.ProcessText(text, metadata)
	if err != nil {
		return err
	}

	return r.store.AddDocuments(ctx, docs)
}

func (r *ragService) Search(ctx context.Context, query string) ([]schema.Document, error) {
	return r.store.Search(ctx, query, 3)
}

func (r *ragService) SearchContext(ctx context.Context, query string) (string, error) {
	docs, err := r.store.Search(ctx, query, 3)
	if err != nil {
		return "", err
	}

	return FormatContext(docs), nil
}
