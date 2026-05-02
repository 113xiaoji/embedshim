package pidfd

import "testing"

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

func TestOptionsOverrideValues(t *testing.T) {
	opts, err := newOptions(
		WithMaxEvents(512),
		WithCallbackWorkers(4),
		WithCallbackQueueDepth(8192),
	)
	if err != nil {
		t.Fatalf("newOptions() error = %v", err)
	}
	if opts.maxEvents != 512 {
		t.Fatalf("maxEvents = %d, want 512", opts.maxEvents)
	}
	if opts.callbackWorkers != 4 {
		t.Fatalf("callbackWorkers = %d, want 4", opts.callbackWorkers)
	}
	if opts.callbackQueueDepth != 8192 {
		t.Fatalf("callbackQueueDepth = %d, want 8192", opts.callbackQueueDepth)
	}
}

func TestOptionsRejectInvalidValues(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
	}{
		{name: "max events", opt: WithMaxEvents(-1)},
		{name: "callback workers", opt: WithCallbackWorkers(-1)},
		{name: "callback queue depth", opt: WithCallbackQueueDepth(-1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := newOptions(tc.opt); err == nil {
				t.Fatal("newOptions() error = nil, want validation error")
			}
		})
	}
}

func TestWithMetricsOverridesDefaultMetrics(t *testing.T) {
	metrics := &Metrics{}
	opts, err := newOptions(WithMetrics(metrics))
	if err != nil {
		t.Fatalf("newOptions() error = %v", err)
	}
	if opts.metrics != metrics {
		t.Fatal("metrics option did not install provided metrics")
	}
}

func TestWithWorkerCPUSetsCopiesInput(t *testing.T) {
	cpuSets := [][]int{{0, 1}, {2, 3}}
	opts, err := newOptions(WithWorkerCPUSets(cpuSets))
	if err != nil {
		t.Fatalf("newOptions() error = %v", err)
	}

	cpuSets[0][0] = 99

	if got := opts.workerCPUSets[0][0]; got != 0 {
		t.Fatalf("workerCPUSets[0][0] = %d, want copied value 0", got)
	}
}
