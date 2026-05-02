package embedshim

import (
	"github.com/fuweid/embedshim/pkg/topology"

	"github.com/sirupsen/logrus"
)

const (
	armAffinityModeOff         = "off"
	armAffinityModeRuntimeOnly = "runtime-only"
	armAffinityShardingNUMA    = "numa"
)

func pidfdWorkerCPUSetsFromConfig(cfg *Config) [][]int {
	if cfg == nil {
		return nil
	}
	if cfg.ARMAffinityMode != armAffinityModeRuntimeOnly {
		return nil
	}
	if !cfg.ARMAffinityBindWorkers {
		return nil
	}
	if cfg.ARMAffinitySharding != "" && cfg.ARMAffinitySharding != armAffinityShardingNUMA {
		return nil
	}

	sysfsRoot := cfg.TopologySysfsRoot
	if sysfsRoot == "" {
		sysfsRoot = "/sys"
	}

	nodes, err := topology.LoadNUMANodes(sysfsRoot)
	if err != nil {
		logrus.WithError(err).Warn("failed to load numa topology for pidfd worker affinity")
		return nil
	}

	cpuSets := topology.ShardCPUSets(topology.NUMANodeShards(nodes))
	if len(cpuSets) == 0 {
		logrus.Warn("numa topology has no cpu sets for pidfd worker affinity")
	}
	return cpuSets
}
