package rag

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type DocumentProcessor struct {
	splitter textsplitter.TextSplitter
}

func NewDocumentProcessor(chunkSize, chunkOverlap int) *DocumentProcessor {
	if chunkSize == 0 {
		chunkSize = 1000
	}
	if chunkOverlap == 0 {
		chunkOverlap = 200
	}

	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(chunkOverlap),
		textsplitter.WithSeparators([]string{"\n", " "}),
	)

	return &DocumentProcessor{splitter: splitter}
}

func (p *DocumentProcessor) ProcessFile(filePath string) ([]schema.Document, error) {
	cleanPath := filepath.Clean(filePath)
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", cleanPath, err)
	}

	return p.processContent(string(content), map[string]any{
		"source":    cleanPath,
		"file_type": filepath.Ext(cleanPath),
	})
}

func (p *DocumentProcessor) ProcessDirectory(dirPath string) ([]schema.Document, error) {
	var allDocs []schema.Document

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".md" && ext != ".txt" && ext != ".yaml" && ext != ".yml" {
			return nil
		}

		docs, err := p.ProcessFile(path)
		if err != nil {
			fmt.Printf("Warning: failed to process %s: %v\n", path, err)
			return nil
		}

		allDocs = append(allDocs, docs...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return allDocs, nil
}

func (p *DocumentProcessor) ProcessURL(urlStr string) ([]schema.Document, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, fmt.Errorf("invalid or unsafe URL")
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// #nosec G107
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", parsedURL.String(), err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch URL: status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return p.processContent(string(content), map[string]any{
		"source": parsedURL.String(),
		"type":   "url",
	})
}

func (p *DocumentProcessor) ProcessText(text string, metadata map[string]any) ([]schema.Document, error) {
	return p.processContent(text, metadata)
}

func (p *DocumentProcessor) processContent(content string, metadata map[string]any) ([]schema.Document, error) {
	docs, err := textsplitter.CreateDocuments(p.splitter, []string{content}, []map[string]any{metadata})
	if err != nil {
		return nil, fmt.Errorf("failed to split documents: %w", err)
	}

	return docs, nil
}

func (p *DocumentProcessor) SplitText(text string) ([]string, error) {
	return p.splitter.SplitText(text)
}

func IsValidURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}
