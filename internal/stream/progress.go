package stream

import (
	"fmt"
	"sync"
	"time"
)

// ProgressIndicator displays an animated spinner while streaming is in progress.
type ProgressIndicator struct {
	message string
	ticker  *time.Ticker
	done    chan struct{}
	mu      sync.Mutex
	active  bool
}

// Spinner characters: Unicode Braille pattern for smooth animation
var spinnerChars = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

// NewProgressIndicator creates a new progress indicator with the given message.
func NewProgressIndicator(message string) *ProgressIndicator {
	return &ProgressIndicator{
		message: message,
		done:    make(chan struct{}),
	}
}

// Start begins the spinner animation. Thread-safe.
func (p *ProgressIndicator) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active {
		return
	}

	p.active = true
	p.ticker = time.NewTicker(100 * time.Millisecond)

	go p.run()
}

// run executes the animation loop in a separate goroutine.
func (p *ProgressIndicator) run() {
	spinIndex := 0

	for {
		select {
		case <-p.done:
			return
		case <-p.ticker.C:
			char := spinnerChars[spinIndex%len(spinnerChars)]
			// Use carriage return to overwrite the line
			fmt.Printf("\r%c %s", char, p.message)
			spinIndex++
		}
	}
}

// Stop halts the spinner and clears the line. Thread-safe.
func (p *ProgressIndicator) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return
	}

	p.active = false

	if p.ticker != nil {
		p.ticker.Stop()
	}

	// Close the done channel to signal the goroutine to exit
	select {
	case <-p.done:
		// Already closed
	default:
		close(p.done)
	}

	// Clear the spinner line
	fmt.Printf("\r")
}

// IsActive returns whether the spinner is currently running. Thread-safe.
func (p *ProgressIndicator) IsActive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.active
}
