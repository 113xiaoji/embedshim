# Embedshim Analysis Plan

## Goal
Understand what this repository does, identify which kernel capabilities it depends on, and evaluate whether Kunpeng/microarchitecture-aware work could create a competitive advantage.

## Phases
- [complete] Phase 1: Map repository purpose and entry points.
- [complete] Phase 2: Trace runtime flow and kernel/BPF integration.
- [complete] Phase 3: Identify performance/control-plane bottlenecks.
- [complete] Phase 4: Check Kunpeng/ARM64-relevant platform facts.
- [complete] Phase 5: Synthesize recommendations and risks.
- [complete] Phase 6: Draft ARM affinity design and implementation plan.
- [complete] Phase 7: Implement first slice: configurable exitsnoop BPF map capacity.
- [complete] Phase 8: Add pinned BPF map capacity validation and read-only ARM topology parsing foundation.
- [complete] Phase 9: Implement second slice: async pidfd callback dispatch with configurable queue/worker knobs.
- [complete] Phase 10: Add pidfd callback metrics snapshot and expvar publication.
- [complete] Phase 11: Introduce exitMonitor boundary and move exec tracking through monitor methods.
- [complete] Phase 12: Add cgroup/topology shard helpers and opt-in runtime-only ARM worker affinity.
- [complete] Phase 13: Add ARM64 build profile targets and Linux churn/environment validation scripts.
- [complete] Phase 14: Run remote ARM64 verification on Huawei Cloud EulerOS aarch64 host.
- [complete] Phase 15: Install `ctr` on remote ARM64 host and run embed runtime churn validation.

## Errors Encountered
| Error | Attempt | Resolution |
|-------|---------|------------|
| `go test ./...` failed because `go` is not installed/on PATH in this Windows workspace. | Tried local compile/test verification. | Treat as unverified locally; note that the project targets Linux kernel/runtime behavior. |
| `go test ./pkg/exitsnoop -run 'TestOptions|TestApplyMapSizing' -count=1` failed because `go` is not installed/on PATH. | Tried TDD red verification for new tests. | Added tests and implementation, but marked Go verification as environment-blocked rather than passed. |
| `go test ./pkg/exitsnoop -run TestValidateMapMaxEntries -count=1` failed because `go` is not installed/on PATH. | Tried verification for pinned-map capacity helper. | Added tests and implementation, but marked Go verification as environment-blocked rather than passed. |
| `go test ./pkg/topology -count=1` failed because `go` is not installed/on PATH. | Tried verification for CPU set and NUMA sysfs parsing. | Added tests and implementation, but marked Go verification as environment-blocked rather than passed. |
| `go test ./pkg/pidfd -run "TestOptions|TestCallbackDispatcher" -count=1` failed because `go` is not installed/on PATH. | Tried TDD red/focused verification for pidfd options and dispatcher. | Added tests and implementation, but marked Go verification as environment-blocked rather than passed. |
| `go test ./pkg/pidfd -count=1` failed because `go` is not installed/on PATH. | Tried package verification after wiring dispatcher into epoller. | Marked Go verification as environment-blocked rather than passed. |
| `go test ./pkg/topology -count=1` failed again because `go` is not installed/on PATH. | Tried verification for new cgroup and shard helpers. | Added tests and implementation, but marked Go verification as environment-blocked rather than passed. |
| `go test ./...` failed again because `go` is not installed/on PATH. | Tried full verification after completing all rollout slices. | Marked full Go verification as environment-blocked rather than passed. |
| `make -n binaries-arm64-generic` and `make -n binaries-arm64-lse` failed because `make` is not installed/on PATH. | Tried dry-run validation for new ARM64 build profile targets. | Marked Makefile target verification as environment-blocked rather than passed. |
| Remote ARM64 BPF build initially failed because copied Windows checkout materialized `bpf/vmlinux/vmlinux.h` as a text file instead of the expected symlink. | Ran `make -C bpf` on remote HCE aarch64 host. | Recreated `bpf/vmlinux/vmlinux.h -> vmlinux_513.h` in the remote verification copy only. |
| Remote ARM64 BPF build failed with default BiSheng `clang` because it did not expose the BPF target. | Checked available compiler targets. | Used `CLANG=clang-12`, which supports BPF. |
| Remote ARM64 Go test failed with system Go 1.17.3 because dependencies require newer Go/generics support. | Ran tests with `/usr/bin/go`. | Installed temporary Go 1.21.13 under `/tmp/go1.21.13` and reran without replacing system Go. |
| Remote churn benchmark script could not run because `ctr` is missing on the host. | Ran `bash hack/churn-bench.sh`. | Installed official ARM64 `ctr` v1.7.27 to `/usr/local/bin/ctr`; reran churn against an isolated `embedshim-containerd` instance and passed. |
| Remote Docker-managed containerd socket was not suitable as the final churn target. | Pointed `ctr` at `/var/run/docker/containerd/containerd.sock`. | `ctr` could connect, but the image was not in the namespace and the target was Docker's old containerd rather than the tested embed runtime. Started isolated `/tmp/embedshim-run/containerd.sock` from the newly built `bin/embedshim-containerd` instead. |
