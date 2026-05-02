//go:build !linux
// +build !linux

package pidfd

func setCurrentThreadAffinity(cpus []int) error {
	return nil
}
