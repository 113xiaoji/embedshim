package topology

import (
	"fmt"
	"os"
	"path/filepath"
)

type CgroupPlacement struct {
	EffectiveCPUs CPUSet
	EffectiveMems CPUSet
}

func LoadCgroupPlacement(cgroupRoot string) (CgroupPlacement, error) {
	cpus, err := readCPUSetFile(cgroupRoot, "cpuset.cpus.effective", "cpuset.cpus")
	if err != nil {
		return CgroupPlacement{}, fmt.Errorf("read cgroup cpuset cpus: %w", err)
	}
	mems, err := readCPUSetFile(cgroupRoot, "cpuset.mems.effective", "cpuset.mems")
	if err != nil {
		return CgroupPlacement{}, fmt.Errorf("read cgroup cpuset mems: %w", err)
	}
	return CgroupPlacement{
		EffectiveCPUs: cpus,
		EffectiveMems: mems,
	}, nil
}

func readCPUSetFile(root string, names ...string) (CPUSet, error) {
	var lastErr error
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err == nil {
			return ParseCPUSet(string(raw))
		}
		lastErr = err
	}
	return CPUSet{}, lastErr
}
