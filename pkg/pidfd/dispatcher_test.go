package pidfd

import (
	"testing"
	"time"
)

func TestCallbackDispatcherRunsCallback(t *testing.T) {
	metrics := &Metrics{}
	dispatcher := newCallbackDispatcher(1, 1, metrics)
	dispatcher.start()
	defer dispatcher.close()

	done := make(chan struct{})
	if err := dispatcher.dispatch(func() error {
		close(done)
		return nil
	}); err != nil {
		t.Fatalf("dispatch() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("callback was not executed")
	}

	snapshot := dispatcher.metricsSnapshot()
	if snapshot.CallbacksQueued != 1 {
		t.Fatalf("CallbacksQueued = %d, want 1", snapshot.CallbacksQueued)
	}
	if snapshot.CallbacksCompleted != 1 {
		t.Fatalf("CallbacksCompleted = %d, want 1", snapshot.CallbacksCompleted)
	}
}

func TestCallbackDispatcherBlocksWhenQueueIsFull(t *testing.T) {
	dispatcher := newCallbackDispatcher(0, 1)
	defer dispatcher.close()

	if err := dispatcher.dispatch(func() error { return nil }); err != nil {
		t.Fatalf("first dispatch() error = %v", err)
	}

	blocked := make(chan struct{})
	released := make(chan struct{})
	go func() {
		close(blocked)
		_ = dispatcher.dispatch(func() error { return nil })
		close(released)
	}()

	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("second dispatch goroutine did not start")
	}

	select {
	case <-released:
		t.Fatal("dispatch returned before queue space was available")
	case <-time.After(20 * time.Millisecond):
	}

	dispatcher.close()

	select {
	case <-released:
	case <-time.After(time.Second):
		t.Fatal("dispatch did not return after dispatcher close")
	}

	snapshot := dispatcher.metricsSnapshot()
	if snapshot.QueueBackpressure == 0 {
		t.Fatal("QueueBackpressure = 0, want non-zero")
	}
	if snapshot.DispatchErrors == 0 {
		t.Fatal("DispatchErrors = 0, want non-zero")
	}
}
