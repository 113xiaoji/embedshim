#!/usr/bin/env bash
set -euo pipefail

COUNT="${COUNT:-1000}"
RUNTIME="${RUNTIME:-io.containerd.runtime.v1.embed}"
IMAGE="${IMAGE:-docker.io/library/busybox:latest}"
COMMAND="${COMMAND:-true}"
NAMESPACE="${NAMESPACE:-default}"
PREFIX="${PREFIX:-embedshim-churn}"

command -v ctr >/dev/null 2>&1 || {
	echo "ctr is required" >&2
	exit 1
}

start_ns="$(date +%s%N)"
for i in $(seq 1 "${COUNT}"); do
	id="${PREFIX}-${i}"
	ctr --namespace "${NAMESPACE}" run --rm --runtime "${RUNTIME}" "${IMAGE}" "${id}" "${COMMAND}" >/dev/null
done
end_ns="$(date +%s%N)"

elapsed_ns="$((end_ns - start_ns))"
elapsed_ms="$((elapsed_ns / 1000000))"

echo "runtime=${RUNTIME}"
echo "image=${IMAGE}"
echo "count=${COUNT}"
echo "elapsed_ms=${elapsed_ms}"
if [ "${elapsed_ms}" -gt 0 ]; then
	echo "containers_per_second=$((COUNT * 1000 / elapsed_ms))"
fi
