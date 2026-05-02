package exitsnoop

import (
	"testing"

	"github.com/cilium/ebpf"
)

func TestOptionsDefaultMapMaxEntries(t *testing.T) {
	opts, err := newOptions()
	if err != nil {
		t.Fatalf("newOptions() error = %v", err)
	}
	if opts.mapMaxEntries != defaultMapMaxEntries {
		t.Fatalf("mapMaxEntries = %d, want %d", opts.mapMaxEntries, defaultMapMaxEntries)
	}
}

func TestWithMapMaxEntriesOverridesDefault(t *testing.T) {
	opts, err := newOptions(WithMapMaxEntries(32768))
	if err != nil {
		t.Fatalf("newOptions() error = %v", err)
	}
	if opts.mapMaxEntries != 32768 {
		t.Fatalf("mapMaxEntries = %d, want 32768", opts.mapMaxEntries)
	}
}

func TestWithMapMaxEntriesRejectsTooSmallValue(t *testing.T) {
	_, err := newOptions(WithMapMaxEntries(1))
	if err == nil {
		t.Fatal("newOptions() error = nil, want validation error")
	}
}

func TestApplyMapSizingUpdatesKnownMapsOnly(t *testing.T) {
	spec := &ebpf.CollectionSpec{
		Maps: map[string]*ebpf.MapSpec{
			bpfMapTracingTasks: {MaxEntries: 2048},
			bpfMapExitedEvents: {MaxEntries: 2048},
			"unrelated":        {MaxEntries: 7},
		},
	}

	if err := applyMapSizing(spec, 8192); err != nil {
		t.Fatalf("applyMapSizing() error = %v", err)
	}
	if got := spec.Maps[bpfMapTracingTasks].MaxEntries; got != 8192 {
		t.Fatalf("tracing_tasks MaxEntries = %d, want 8192", got)
	}
	if got := spec.Maps[bpfMapExitedEvents].MaxEntries; got != 8192 {
		t.Fatalf("exited_events MaxEntries = %d, want 8192", got)
	}
	if got := spec.Maps["unrelated"].MaxEntries; got != 7 {
		t.Fatalf("unrelated MaxEntries = %d, want 7", got)
	}
}

func TestValidateMapMaxEntriesAcceptsMatchingCapacity(t *testing.T) {
	if err := validateMapMaxEntries("tracing_tasks", 32768, 32768); err != nil {
		t.Fatalf("validateMapMaxEntries() error = %v", err)
	}
}

func TestValidateMapMaxEntriesRejectsMismatchedCapacity(t *testing.T) {
	err := validateMapMaxEntries("tracing_tasks", 2048, 32768)
	if err == nil {
		t.Fatal("validateMapMaxEntries() error = nil, want mismatch error")
	}
}
