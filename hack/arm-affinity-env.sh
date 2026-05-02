#!/usr/bin/env bash
set -euo pipefail

echo "== kernel =="
uname -a

echo
echo "== cpu =="
lscpu || true

echo
echo "== numa cpulists =="
if [ -d /sys/devices/system/node ]; then
	for node in /sys/devices/system/node/node*; do
		[ -d "${node}" ] || continue
		printf '%s ' "$(basename "${node}")"
		cat "${node}/cpulist"
	done
fi

echo
echo "== cgroup cpuset effective =="
for f in /sys/fs/cgroup/cpuset.cpus.effective /sys/fs/cgroup/cpuset.mems.effective; do
	[ -f "${f}" ] && printf '%s=%s\n' "${f}" "$(cat "${f}")"
done
