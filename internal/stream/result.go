package stream

import (
	"fmt"
	"time"
)

// ContentType represents the type of content in a StreamResult.
type ContentType string

const (
	// ContentTypeText represents plain text content.
	ContentTypeText ContentType = "text"
	// ContentTypeJSON represents JSON content.
	ContentTypeJSON ContentType = "json"
	// ContentTypeError represents an error message.
	ContentTypeError ContentType = "error"
	// ContentTypeProgress represents a progress indicator.
	ContentTypeProgress ContentType = "progress"
)

// StreamResult represents a single immutable chunk of AI-generated content.
type StreamResult struct {
	SequenceNumber int
	Content        []byte
	ReceivedAt     time.Time
	Size           int
	ContentType    ContentType
}

// NewStreamResult creates a new StreamResult chunk.
// Returns error if validation fails.
func NewStreamResult(sequenceNumber int, content []byte, contentType ContentType) (*StreamResult, error) {
	if sequenceNumber < 1 {
		return nil, fmt.Errorf("sequence number must be >= 1, got %d", sequenceNumber)
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("content cannot be empty")
	}

	if contentType == "" {
		contentType = ContentTypeText
	}

	// Validate content type
	switch contentType {
	case ContentTypeText, ContentTypeJSON, ContentTypeError, ContentTypeProgress:
		// Valid content types
	default:
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}

	return &StreamResult{
		SequenceNumber: sequenceNumber,
		Content:        content,
		ReceivedAt:     time.Now(),
		Size:           len(content),
		ContentType:    contentType,
	}, nil
}

// String returns a string representation of the result.
func (sr *StreamResult) String() string {
	return fmt.Sprintf("StreamResult{seq=%d, type=%s, size=%d, time=%s}",
		sr.SequenceNumber, sr.ContentType, sr.Size, sr.ReceivedAt.Format(time.RFC3339Nano))
}
