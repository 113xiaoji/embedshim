# Embedshim Analysis Findings

## Repository Findings
- README says embedshim is a containerd task runtime implementation/plugin. Its goal is daemonless/low-overhead container lifecycle management by avoiding a per-pod/per-container Go shim process.
- Build outputs are `embedshim-containerd` and `embedshim-runcext`. `embedshim-containerd` embeds the plugin into containerd; `embedshim-runcext` appears to be a runc wrapper/helper.
- Explicit kernel requirements from README: raw tracepoint BPF >= 4.18, CO-RE/BTF vmlinux >= 5.4, pidfd polling >= 5.3.
- Module is old containerd-era code: Go 1.16, containerd v1.5.13, cilium/ebpf v0.12.2, runc v1.1.2.
- `cmd/embedshim-containerd/main.go` starts a normal containerd binary with this repository imported for plugin registration.
- `plugin.go` registers runtime plugin `io.containerd.runtime.v1.embed`, initializes a TaskManager, starts/pins exitsnoop BPF, creates a trace ID allocator, and reloads existing tasks from state.
- `shim.go` implements containerd runtime.Task behavior in-process. It creates OCI bundles, invokes runc, tracks init/exec process state, exposes cgroup stats, and avoids a separate shim daemon.
- `cmd/embedshim-runcext` wraps `runc exec` so the parent can receive the exec pid/status early enough to install pidfd/BPF monitoring.

## Kernel / BPF Findings
- BPF is compiled before Go generation in the Makefile. The project carries `bpf/exitsnoop.bpf.c`, libbpf as a submodule, and generated/load code under `pkg/exitsnoop`.
- `bpf/exitsnoop.bpf.c` attaches to raw tracepoint `sched_process_exit`, filters main thread exits (`pid == tid`), checks pid namespace identity, reads `task_struct.exit_code`, and writes exit status keyed by userspace trace ID.
- `monitor.go` opens pidfds for init processes, registers them in an epoll loop, and on pidfd readability looks up the BPF exit event to set the container exit status.
- `exec_process.go` uses a runcext synchronization pipe to learn the exec pid, traces it in a separate in-memory BPF store, and epolls the exec pidfd.
- Exit tracking is two-level: pidfd gives race-resistant liveness/termination notification; eBPF gives exit code recovery for non-parent/reconnect cases.
- There is no current Kunpeng, NUMA, affinity, cpuset, PMU, or ARM-specific runtime policy in repository code. The only architecture conditional is the BPF compile target mapping `aarch64 -> arm64` in `bpf/Makefile`.
- Runtime events are incomplete per README (`Task Event(Create/Start/Exit/Delete/OOM) support` unchecked), and `TaskManager.events` is stored but not used for publishing task lifecycle events.
- Known internal TODO/FIXME areas relevant to competitiveness: exec FIFO buffering risk, duplicated BPF attach/pinning paths, unchecked pid poller goroutine return, checkpoint missing, and Wait ignoring context.

## Kunpeng / Microarchitecture Findings
- Huawei Kunpeng docs emphasize Kunpeng 920 as an ARM-based data-center processor and call out NUMA, pipeline, cache/prefetch, Armv8-A/NEON, compiler, and BoostKit as optimization dimensions.
- Official Kunpeng Docker tuning docs recommend binding containers to CPUs and memory nodes to maximize container performance; for Kunpeng 920 5250 example, NUMA nodes map to CPU ranges 0-23, 24-47, 48-71, and 72-95.
- Kunpeng docs specifically warn that cross-die/cross-chip memory access and CPU competition can degrade performance, including L1 TLB flush/miss effects.
- Linux cgroup v2 exposes `cpuset.cpus` and `cpuset.mems`, including effective CPU/memory-node views and cpuset partition/isolated partition concepts. This is a natural integration point for runtime-level Kunpeng affinity.
- Kunpeng DevKit provides Performance Boundary Analyzer/System Profiler covering cache misses, memory access, NUMA, microarchitecture, miss latencies, hotspot functions, CPU usage, softirqs, and other system metrics.
- Current containerd 2.x removed runtime v1/runc v1 shims and stabilized the sandbox service. This repository is still built around containerd v1.5/runtime v1 plugin shape, so a productized direction likely needs migration to runtime v2/sandbox or external plugin integration.
- The pidfd epoll loop now has a bounded callback dispatcher path in code, but local validation is still blocked until a Linux/Go environment can compile and run `pkg/pidfd` tests.
- Runtime-only ARM affinity is implemented as an opt-in pidfd callback worker binding policy. It discovers NUMA CPU sets from sysfs and never changes workload cpuset/memset.
- Metrics are available as a standard expvar snapshot under `embedshim_pidfd`; Prometheus/containerd-native metric registration remains a later integration choice.

## Open Questions
- Need a Linux ARM64/Kunpeng host or CI worker with Go installed to compile and run the eBPF/pidfd paths. The current Windows workspace can inspect and edit code, but cannot validate kernel behavior.
- Need product decision on whether `bpf_map_max_entries` should be hot-changeable. The current implementation deliberately rejects a mismatch against already-pinned BPF maps and asks the operator to remove pinned exitsnoop objects or keep the previous capacity.
- Need Linux/ARM64 benchmark results to tune the default values for `bpf_map_max_entries`, `epoll_batch_size`, `pidfd_callback_workers`, and `pidfd_callback_queue_depth`.

## Remote ARM64 Verification Findings
- Remote compile/unit verification was completed on `124.70.162.35`, a Huawei Cloud EulerOS 2.0 `aarch64` host running Linux 5.10.
- The remote host's system Go 1.17.3 is too old for the repository dependency graph; Go 1.21.13 was installed temporarily under `/tmp/go1.21.13` and used for validation without replacing the host Go.
- BPF build on the remote host requires `CLANG=clang-12`; the default BiSheng clang did not expose the BPF target.
- `make -C bpf CLANG=clang-12 LLVM_STRIP=llvm-strip` and `go generate ./...` passed after recreating the `bpf/vmlinux/vmlinux.h` symlink in the remote verification copy.
- `go test ./... -count=1` passed remotely with Go 1.21.13, covering the new exitsnoop options, pidfd dispatcher/options/metrics path, and topology helpers.
- `make binaries-arm64-generic CLANG=clang-12` and `make binaries-arm64-lse CLANG=clang-12` both passed remotely and produced ARM aarch64 ELF binaries.
- `hack/arm-affinity-env.sh` ran successfully and reported ARM64 topology/feature information.
- `ctr` was installed on the remote host as `/usr/local/bin/ctr` from the official ARM64 containerd v1.7.27 release; system `/usr/bin/containerd` was not replaced.
- Docker's containerd socket at `/var/run/docker/containerd/containerd.sock` was reachable with `ctr`, but the final churn run used an isolated `embedshim-containerd` instance from the freshly built repository binary so the runtime under test was `io.containerd.runtime.v1.embed`.
- The isolated embedshim daemon loaded `io.containerd.runtime.v1.embed`, imported `quay.io/almalinuxorg/9-base:9.4`, and completed `hack/churn-bench.sh` with `COUNT=1000`, `/bin/true`, and the embed runtime in 52497 ms, reporting 19 containers/s.
- Churn log review found only the expected CRI CNI init warning for the temporary daemon and repeated runtime-v1 deprecation warnings; no churn-failing error was observed.

## Proposed Optimization Plan
- Phase A: Make capacity configurable. Add runtime config for BPF map max entries, epoll batch size, callback worker count, and queue depth; build BPF object from config or ship preset objects such as small/default/large.
- Phase B: Unify exit monitoring. Replace separate init pinned store and exec in-memory store with one `ExitMonitor` abstraction, one tracepoint attach, one trace ID namespace, and a value schema that records process kind, namespace, container ID, exec ID, pid namespace, and optional NUMA shard.
- Phase C: Decouple epoll from callback execution. Keep `epoll_wait` loop single-purpose: drain pidfd events quickly, close/remove fd, enqueue `ExitEvent`. Process callbacks in bounded worker pools with backpressure, error logging, and metrics.
- Phase D: Add NUMA-aware sharding for Kunpeng. Discover node CPU/memory topology from sysfs/cgroup, assign monitored tasks to shards, bind worker goroutines/OS threads to CPU sets when enabled, and align cgroup `cpuset.cpus`/`cpuset.mems` policy with runtime-side event workers.
- Phase E: Add observability and SLOs. Track trace registration failures, map occupancy, missing BPF exit status, epoll drain latency, callback queue latency, callback processing time, pidfd events per second, and P99/P999 container exit notification latency.
- Phase F: Validate with stress workload. Test short-lived container churn, exec churn, containerd restart recovery, map capacity exhaustion, and Kunpeng NUMA/cpuset profiles before considering microarchitecture-specific tuning.
