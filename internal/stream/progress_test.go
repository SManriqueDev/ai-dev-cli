package stream

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProgressIndicator_Start(t *testing.T) {
	pi := NewProgressIndicator("Testing...")
	pi.Start()
	defer pi.Stop()

	require.True(t, pi.IsActive(), "spinner should be active after Start()")
}

func TestProgressIndicator_Stop(t *testing.T) {
	pi := NewProgressIndicator("Testing...")
	pi.Start()
	time.Sleep(100 * time.Millisecond)

	pi.Stop()

	require.False(t, pi.IsActive(), "spinner should not be active after Stop()")
}

func TestProgressIndicator_StartStop_Multiple(t *testing.T) {
	pi := NewProgressIndicator("Testing...")

	// Start and stop multiple times
	for i := 0; i < 3; i++ {
		pi.Start()
		require.True(t, pi.IsActive())

		time.Sleep(50 * time.Millisecond)

		pi.Stop()
		require.False(t, pi.IsActive())

		time.Sleep(50 * time.Millisecond)
	}
}

func TestProgressIndicator_StopWithoutStart(t *testing.T) {
	pi := NewProgressIndicator("Testing...")

	// Should not panic when stopping without starting
	require.NotPanics(t, func() {
		pi.Stop()
	})

	require.False(t, pi.IsActive())
}

func TestProgressIndicator_ConcurrentStartStop(t *testing.T) {
	pi := NewProgressIndicator("Concurrent test...")
	var wg sync.WaitGroup

	// Launch multiple goroutines trying to start/stop concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pi.Start()
			time.Sleep(50 * time.Millisecond)
			pi.Stop()
		}()
	}

	wg.Wait()

	// Final state should be inactive
	require.False(t, pi.IsActive())
}

func TestProgressIndicator_IsActive(t *testing.T) {
	pi := NewProgressIndicator("Testing...")

	require.False(t, pi.IsActive(), "should not be active initially")

	pi.Start()
	require.True(t, pi.IsActive(), "should be active after Start()")

	pi.Stop()
	require.False(t, pi.IsActive(), "should not be active after Stop()")
}

func TestProgressIndicator_DoubleSta(t *testing.T) {
	pi := NewProgressIndicator("Testing...")

	// Starting twice should not cause issues
	pi.Start()
	pi.Start() // Second start should be idempotent

	require.True(t, pi.IsActive())

	pi.Stop()
	require.False(t, pi.IsActive())
}

func TestProgressIndicator_DoubleStop(t *testing.T) {
	pi := NewProgressIndicator("Testing...")

	pi.Start()
	pi.Stop()
	pi.Stop() // Second stop should not panic

	require.False(t, pi.IsActive())
}
