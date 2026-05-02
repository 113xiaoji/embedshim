# PIDFD Async Callbacks Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Decouple pidfd `epoll_wait` draining from exit callback execution with a bounded worker queue.

**Architecture:** Keep `pkg/pidfd.Epoller` as the public abstraction and preserve `NewEpoller()` compatibility by adding option arguments. `Run()` removes closed pidfds from epoll and closes the fd quickly, then enqueues the callback into a bounded dispatcher. Plugin config exposes batch size, queue depth, and worker count for high-churn ARM/Kunpeng hosts.

**Tech Stack:** Go 1.16-era code, Linux pidfd, epoll, `golang.org/x/sys/unix`, containerd runtime plugin config.

---

### Task 1: Add pidfd epoller options and dispatcher tests

**Files:**
- Create: `pkg/pidfd/options.go`
- Create: `pkg/pidfd/options_test.go`
- Create: `pkg/pidfd/dispatcher_test.go`

- [x] **Step 1: Write option tests**

Create tests for default values, overrides, and too-small rejection:

```go
func TestOptionsDefaultValues(t *testing.T) {
	opts, err := newOptions()
	if err != nil {
		t.Fatalf("newOptions() error = %v", err)
	}
	if opts.maxEvents != defaultMaxEvents {
		t.Fatalf("maxEvents = %d, want %d", opts.maxEvents, defaultMaxEvents)
	}
	if opts.callbackWorkers != defaultCallbackWorkers {
		t.Fatalf("callbackWorkers = %d, want %d", opts.callbackWorkers, defaultCallbackWorkers)
	}
	if opts.callbackQueueDepth != defaultCallbackQueueDepth {
		t.Fatalf("callbackQueueDepth = %d, want %d", opts.callbackQueueDepth, defaultCallbackQueueDepth)
	}
}
```

- [x] **Step 2: Write dispatcher tests**

Create tests proving callbacks run through workers and queue backpressure blocks when the queue is full:

```go
func TestCallbackDispatcherRunsCallback(t *testing.T) {
	dispatcher := newCallbackDispatcher(1, 1)
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
}
```

- [x] **Step 3: Run tests to verify they fail**

Run: `go test ./pkg/pidfd -run 'TestOptions|TestCallbackDispatcher' -count=1`

Expected: FAIL because options and dispatcher helpers do not exist yet.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

### Task 2: Implement bounded callback dispatcher

**Files:**
- Create: `pkg/pidfd/options.go`
- Create: `pkg/pidfd/dispatcher.go`

- [x] **Step 1: Implement options**

Add `Option`, `newOptions`, `WithMaxEvents`, `WithCallbackWorkers`, and `WithCallbackQueueDepth`. A value of `0` keeps the default; values smaller than `1` after defaulting are rejected.

- [x] **Step 2: Implement dispatcher**

Add a dispatcher with:

```go
type callbackDispatcher struct {
	queue chan pidOnClose
	done  chan struct{}
	once  sync.Once
	wg    sync.WaitGroup
}
```

`dispatch` must block when the queue is full and return an error if the dispatcher is closed. Worker callbacks keep the current behavior of ignoring callback return errors.

- [x] **Step 3: Run focused tests**

Run: `go test ./pkg/pidfd -run 'TestOptions|TestCallbackDispatcher' -count=1`

Expected: PASS.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

### Task 3: Wire dispatcher into Epoller

**Files:**
- Modify: `pkg/pidfd/epoll.go`

- [x] **Step 1: Update `NewEpoller` signature**

Change `NewEpoller()` to `NewEpoller(opts ...Option)` and initialize max events plus dispatcher.

- [x] **Step 2: Update `Run`**

`Run()` should remove the fd from epoll, remove the callback from the map, close the fd, and call `dispatcher.dispatch(onClose)` instead of invoking `onClose()` inline.

- [x] **Step 3: Update `Close`**

`Close()` should close the epoll fd and close the dispatcher once.

- [x] **Step 4: Run pidfd tests**

Run: `go test ./pkg/pidfd -count=1`

Expected: PASS.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

### Task 4: Expose runtime config

**Files:**
- Modify: `plugin.go`
- Modify: `monitor.go`
- Modify: `docs/superpowers/specs/2026-05-02-arm-affinity-runtime-design.md`

- [x] **Step 1: Extend plugin config**

Add:

```go
EpollBatchSize         int `toml:"epoll_batch_size"`
PIDFDCallbackWorkers  int `toml:"pidfd_callback_workers"`
PIDFDCallbackQueueDepth int `toml:"pidfd_callback_queue_depth"`
```

- [x] **Step 2: Pass options to pidfd epoller**

Call `pidfd.NewEpoller` with `WithMaxEvents`, `WithCallbackWorkers`, and `WithCallbackQueueDepth`.

- [x] **Step 3: Document config**

Update the ARM design doc config sample to use the concrete plugin keys.

- [x] **Step 4: Run full tests**

Run: `go test ./...`

Expected: PASS.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.
