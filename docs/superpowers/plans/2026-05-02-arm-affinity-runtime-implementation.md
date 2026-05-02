# ARM Affinity Runtime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first safe slice of ARM affinity support by making exit-monitor BPF capacity configurable and preparing the exit-monitor code for later unified/sharded processing.

**Architecture:** Start with configuration and pure helper tests before changing runtime behavior. Preserve existing public call sites by adding option-based functions while keeping compatibility wrappers. Later tasks can move init/exec to a unified `ExitMonitor` without changing the BPF object format again.

**Tech Stack:** Go 1.16-era code, containerd v1.5 runtime plugin, cilium/ebpf, raw tracepoint BPF, pidfd/epoll.

---

### Task 1: Configurable exitsnoop BPF map capacity

**Files:**
- Create: `pkg/exitsnoop/options.go`
- Create: `pkg/exitsnoop/options_test.go`
- Modify: `pkg/exitsnoop/load.go`
- Modify: `plugin.go`
- Modify: `monitor.go`

- [x] **Step 1: Write failing tests for exitsnoop options**

Create `pkg/exitsnoop/options_test.go` with tests that define the desired helper behavior:

```go
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
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/exitsnoop -run 'TestOptions|TestApplyMapSizing' -count=1`

Expected: FAIL because `newOptions`, `WithMapMaxEntries`, `defaultMapMaxEntries`, and `applyMapSizing` do not exist yet.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

- [x] **Step 3: Add minimal options implementation**

Create `pkg/exitsnoop/options.go`:

```go
package exitsnoop

import (
	"fmt"

	"github.com/cilium/ebpf"
)

const (
	defaultMapMaxEntries uint32 = 2048
	minMapMaxEntries     uint32 = 16
)

type options struct {
	mapMaxEntries uint32
}

type Option func(*options) error

func newOptions(opts ...Option) (*options, error) {
	out := &options{
		mapMaxEntries: defaultMapMaxEntries,
	}
	for _, opt := range opts {
		if err := opt(out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func WithMapMaxEntries(entries uint32) Option {
	return func(o *options) error {
		if entries == 0 {
			return nil
		}
		if entries < minMapMaxEntries {
			return fmt.Errorf("bpf map max entries %d is smaller than minimum %d", entries, minMapMaxEntries)
		}
		o.mapMaxEntries = entries
		return nil
	}
}

func applyMapSizing(spec *ebpf.CollectionSpec, entries uint32) error {
	for _, name := range []string{bpfMapTracingTasks, bpfMapExitedEvents} {
		m, ok := spec.Maps[name]
		if !ok {
			return fmt.Errorf("bpf map %s is missing from collection spec", name)
		}
		m.MaxEntries = entries
	}
	return nil
}
```

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/exitsnoop -run 'TestOptions|TestApplyMapSizing' -count=1`

Expected: PASS.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

- [x] **Step 5: Wire options into BPF load paths**

Modify `pkg/exitsnoop/load.go`:

```go
func NewStoreFromAttach(opts ...Option) (_ *Store, retErr error) {
	options, err := newOptions(opts...)
	if err != nil {
		return nil, err
	}

	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(progByteCode))
	if err != nil {
		return nil, err
	}
	if err := applyMapSizing(spec, options.mapMaxEntries); err != nil {
		return nil, err
	}
	...
}

func EnsureRunning(bpffsRoot string) error {
	return EnsureRunningWithOptions(bpffsRoot)
}

func EnsureRunningWithOptions(bpffsRoot string, opts ...Option) error {
	options, err := newOptions(opts...)
	if err != nil {
		return err
	}
	...
	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(progByteCode))
	if err != nil {
		return err
	}
	if err := applyMapSizing(spec, options.mapMaxEntries); err != nil {
		return err
	}
	...
}
```

- [x] **Step 6: Expose containerd plugin config**

Modify `plugin.go`:

```go
type Config struct {
	BPFMapMaxEntries uint32 `toml:"bpf_map_max_entries"`
}
```

Modify `TaskManager.init()` to pass the configured value:

```go
err := exitsnoop.EnsureRunningWithOptions(
	manager.rootDir,
	exitsnoop.WithMapMaxEntries(manager.config.BPFMapMaxEntries),
)
```

Modify `newMonitor` signature and call site so the exec attach path receives the same sizing:

```go
manager.monitor, err = newMonitor(manager.rootDir, manager.config)
```

```go
func newMonitor(stateDir string, cfg *Config) (_ *monitor, retErr error) {
	...
	execStore, err := exitsnoop.NewStoreFromAttach(
		exitsnoop.WithMapMaxEntries(cfg.BPFMapMaxEntries),
	)
	...
}
```

- [x] **Step 7: Run focused tests**

Run: `go test ./pkg/exitsnoop -count=1`

Expected: PASS.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

- [x] **Step 8: Run full tests**

Run: `go test ./...`

Expected: PASS.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

### Task 2: Document first configuration surface

**Files:**
- Modify: `docs/superpowers/specs/2026-05-02-arm-affinity-runtime-design.md`

- [x] **Step 1: Add config example**

Add a short example showing:

```toml
[plugins."io.containerd.runtime.v1.embed"]
bpf_map_max_entries = 32768
```

- [x] **Step 2: Run markdown sanity check**

Run: `rg -n "TBD|TODO|FIXME" docs/superpowers/specs/2026-05-02-arm-affinity-runtime-design.md`

Expected: no matches.

### Task 3: Validate existing pinned BPF map capacity

**Files:**
- Modify: `pkg/exitsnoop/options.go`
- Modify: `pkg/exitsnoop/options_test.go`
- Modify: `pkg/exitsnoop/load.go`

- [x] **Step 1: Write failing tests for capacity mismatch**

Extend `pkg/exitsnoop/options_test.go`:

```go
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
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/exitsnoop -run TestValidateMapMaxEntries -count=1`

Expected: FAIL because `validateMapMaxEntries` does not exist yet.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

- [x] **Step 3: Implement validation helper**

Add to `pkg/exitsnoop/options.go`:

```go
func validateMapMaxEntries(name string, got, want uint32) error {
	if got == want {
		return nil
	}
	return fmt.Errorf("pinned bpf map %s has max entries %d, want %d; remove the pinned exitsnoop objects or keep the previous capacity", name, got, want)
}
```

- [x] **Step 4: Validate pinned maps when program already exists**

Add to `pkg/exitsnoop/load.go`:

```go
func validatePinnedMapSizing(rootDir string, entries uint32) error {
	for _, name := range []string{bpfMapTracingTasks, bpfMapExitedEvents} {
		m, err := loadPinnedMap(filepath.Join(rootDir, name))
		if err != nil {
			return err
		}
		info, err := m.Info()
		m.Close()
		if err != nil {
			return err
		}
		if err := validateMapMaxEntries(name, info.MaxEntries, entries); err != nil {
			return err
		}
	}
	return nil
}
```

Change the existing pinned-program fast path:

```go
if err == nil {
	return validatePinnedMapSizing(rootDir, options.mapMaxEntries)
}
```

- [x] **Step 5: Run focused tests**

Run: `go test ./pkg/exitsnoop -run 'TestOptions|TestApplyMapSizing|TestValidateMapMaxEntries' -count=1`

Expected: PASS.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

### Task 4: Add read-only ARM topology parsing foundation

**Files:**
- Create: `pkg/topology/cpuset.go`
- Create: `pkg/topology/topology.go`
- Create: `pkg/topology/cpuset_test.go`
- Create: `pkg/topology/topology_test.go`

- [x] **Step 1: Write failing tests for CPU set parsing**

Create `pkg/topology/cpuset_test.go`:

```go
package topology

import (
	"reflect"
	"testing"
)

func TestParseCPUSetExpandsRangesAndSingles(t *testing.T) {
	set, err := ParseCPUSet("0-3,8,10-11")
	if err != nil {
		t.Fatalf("ParseCPUSet() error = %v", err)
	}
	want := []int{0, 1, 2, 3, 8, 10, 11}
	if got := set.List(); !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %v, want %v", got, want)
	}
	if got := set.String(); got != "0-3,8,10-11" {
		t.Fatalf("String() = %q, want %q", got, "0-3,8,10-11")
	}
}

func TestParseCPUSetRejectsDescendingRange(t *testing.T) {
	if _, err := ParseCPUSet("4-2"); err == nil {
		t.Fatal("ParseCPUSet() error = nil, want descending range error")
	}
}
```

- [x] **Step 2: Write failing tests for NUMA node loading**

Create `pkg/topology/topology_test.go`:

```go
package topology

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadNUMANodesReadsAndSortsNodeCPULists(t *testing.T) {
	root := t.TempDir()
	nodeRoot := filepath.Join(root, "devices", "system", "node")
	for _, name := range []string{"node1", "node0", "not-a-node"} {
		if err := os.MkdirAll(filepath.Join(nodeRoot, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(nodeRoot, "node1", "cpulist"), []byte("4-7\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeRoot, "node0", "cpulist"), []byte("0-3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	nodes, err := LoadNUMANodes(root)
	if err != nil {
		t.Fatalf("LoadNUMANodes() error = %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("len(nodes) = %d, want 2", len(nodes))
	}
	if nodes[0].ID != 0 || nodes[1].ID != 1 {
		t.Fatalf("node IDs = %d,%d, want 0,1", nodes[0].ID, nodes[1].ID)
	}
	if got, want := nodes[1].CPUs.List(), []int{4, 5, 6, 7}; !reflect.DeepEqual(got, want) {
		t.Fatalf("node1 CPUs = %v, want %v", got, want)
	}
}
```

- [x] **Step 3: Run tests to verify they fail**

Run: `go test ./pkg/topology -count=1`

Expected: FAIL because the topology package does not exist yet.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.

- [x] **Step 4: Implement CPUSet parser**

Create `pkg/topology/cpuset.go` with a parser for Linux cpulist strings such as `0-3,8,10-11`.

- [x] **Step 5: Implement NUMA sysfs loader**

Create `pkg/topology/topology.go` with `LoadNUMANodes(sysfsRoot string)`.

- [x] **Step 6: Run topology tests**

Run: `go test ./pkg/topology -count=1`

Expected: PASS.

Actual: attempted, but blocked because `go` is not installed/on PATH in this workspace.
