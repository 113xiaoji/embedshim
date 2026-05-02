//go:build linux
// +build linux

package pidfd

import "golang.org/x/sys/unix"

func setCurrentThreadAffinity(cpus []int) error {
	var set unix.CPUSet
	set.Zero()
	for _, cpu := range cpus {
		set.Set(cpu)
	}
	return unix.SchedSetaffinity(0, &set)
}
