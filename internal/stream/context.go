package stream

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// StreamStatus represents the status of a streaming operation.
type StreamStatus string

const (
	// StatusRunning indicates the stream is currently active.
	StatusRunning StreamStatus = "running"
	// StatusCompleted indicates the stream finished successfully.
	StatusCompleted StreamStatus = "completed"
	// StatusInterrupted indicates the stream was interrupted by user.
	StatusInterrupted StreamStatus = "interrupted"
	// StatusFailed indicates the stream encountered an error.
	StatusFailed StreamStatus = "failed"
)

// StreamContext manages the lifecycle and metadata for a streaming operation.
type StreamContext struct {
	mu            sync.RWMutex
	OperationID   string
	StartedAt     time.Time
	Command       string
	FilePath      string
	Status        StreamStatus
	ChunkCount    int
	BytesReceived int64
	InterruptedAt *time.Time
	ErrorMessage  string
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewStreamContext creates a new StreamContext for a streaming operation.
func NewStreamContext(command, filePath string) *StreamContext {
	ctx, cancel := context.WithCancel(context.Background())

	return &StreamContext{
		OperationID: fmt.Sprintf("op_%d", time.Now().UnixNano()),
		StartedAt:   time.Now(),
		Command:     command,
		FilePath:    filePath,
		Status:      StatusRunning,
		ChunkCount:  0,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Context returns the underlying context for this streaming operation.
func (sc *StreamContext) Context() context.Context {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.ctx
}

// Cancel cancels the streaming operation.
func (sc *StreamContext) Cancel() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.cancel != nil {
		sc.cancel()
	}
}

// RecordChunk records that a new chunk was received.
func (sc *StreamContext) RecordChunk(size int) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.Status == StatusRunning {
		sc.ChunkCount++
		sc.BytesReceived += int64(size)
	}
}

// CompleteSuccessfully marks the operation as completed successfully.
func (sc *StreamContext) CompleteSuccessfully() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.Status == StatusRunning {
		sc.Status = StatusCompleted
	}
}

// Interrupt marks the operation as interrupted by user (Ctrl+C).
func (sc *StreamContext) Interrupt() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.Status == StatusRunning {
		sc.Status = StatusInterrupted
		now := time.Now()
		sc.InterruptedAt = &now
		if sc.cancel != nil {
			sc.cancel()
		}
	}
}

// RecordError marks the operation as failed with an error message.
func (sc *StreamContext) RecordError(errMsg string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.Status == StatusRunning {
		sc.Status = StatusFailed
		sc.ErrorMessage = errMsg
	}
}

// GetSnapshot returns a read-only snapshot of the current state.
func (sc *StreamContext) GetSnapshot() StreamContextSnapshot {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	snapshot := StreamContextSnapshot{
		OperationID:   sc.OperationID,
		StartedAt:     sc.StartedAt,
		Command:       sc.Command,
		FilePath:      sc.FilePath,
		Status:        sc.Status,
		ChunkCount:    sc.ChunkCount,
		BytesReceived: sc.BytesReceived,
		ErrorMessage:  sc.ErrorMessage,
	}

	if sc.InterruptedAt != nil {
		snapshot.InterruptedAt = *sc.InterruptedAt
		snapshot.HasInterruptedAt = true
	}

	return snapshot
}

// StreamContextSnapshot is a read-only snapshot of a StreamContext.
type StreamContextSnapshot struct {
	OperationID      string
	StartedAt        time.Time
	Command          string
	FilePath         string
	Status           StreamStatus
	ChunkCount       int
	BytesReceived    int64
	InterruptedAt    time.Time
	HasInterruptedAt bool
	ErrorMessage     string
}

// InterruptHandler manages graceful shutdown when user sends Ctrl+C.
type InterruptHandler struct {
	mu           sync.Mutex
	streamCtx    *StreamContext
	signalChan   chan os.Signal
	cancelFuncs  []context.CancelFunc
	cleanupFuncs []func() error
	stopOnce     sync.Once
}

// NewInterruptHandler creates a new InterruptHandler for managing signals.
func NewInterruptHandler(streamCtx *StreamContext) *InterruptHandler {
	handler := &InterruptHandler{
		streamCtx:    streamCtx,
		signalChan:   make(chan os.Signal, 1),
		cancelFuncs:  make([]context.CancelFunc, 0),
		cleanupFuncs: make([]func() error, 0),
	}

	// Register for SIGINT (Ctrl+C)
	signal.Notify(handler.signalChan, os.Interrupt, syscall.SIGINT)

	// Start goroutine to handle signals
	go handler.waitForSignal()

	return handler
}

// RegisterCancelFunc registers a cancel function to be called on interrupt.
func (ih *InterruptHandler) RegisterCancelFunc(cancel context.CancelFunc) {
	ih.mu.Lock()
	defer ih.mu.Unlock()
	if cancel != nil {
		ih.cancelFuncs = append(ih.cancelFuncs, cancel)
	}
}

// RegisterCleanupFunc registers a cleanup function to be called on interrupt.
func (ih *InterruptHandler) RegisterCleanupFunc(cleanup func() error) {
	ih.mu.Lock()
	defer ih.mu.Unlock()
	if cleanup != nil {
		ih.cleanupFuncs = append(ih.cleanupFuncs, cleanup)
	}
}

// waitForSignal waits for a signal and handles it.
func (ih *InterruptHandler) waitForSignal() {
	<-ih.signalChan
	ih.handleInterrupt()
}

// handleInterrupt handles an interrupt signal.
func (ih *InterruptHandler) handleInterrupt() {
	ih.stopOnce.Do(func() {
		ih.mu.Lock()
		defer ih.mu.Unlock()

		// Display cancellation message immediately
		fmt.Fprintln(os.Stderr, "\nOperation cancelled by user")

		// Mark stream context as interrupted
		ih.streamCtx.Interrupt()

		// Call all registered cancel funcs
		for _, cancel := range ih.cancelFuncs {
			cancel()
		}

		// Call all cleanup functions (fail-safe - don't block on errors)
		for _, cleanup := range ih.cleanupFuncs {
			if cleanup != nil {
				_ = cleanup() // Ignore errors in cleanup
			}
		}

		// Stop receiving signals
		signal.Stop(ih.signalChan)
	})
}

// Stop stops listening for signals and performs cleanup.
func (ih *InterruptHandler) Stop() {
	ih.stopOnce.Do(func() {
		ih.mu.Lock()
		defer ih.mu.Unlock()

		signal.Stop(ih.signalChan)
		close(ih.signalChan)
	})
}
