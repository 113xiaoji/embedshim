package topology

import "testing"

func TestNUMANodeShardsBuildsCPUSetList(t *testing.T) {
	shards := NUMANodeShards([]NUMANode{
		{ID: 0, CPUs: NewCPUSet([]int{0, 1})},
		{ID: 1, CPUs: NewCPUSet([]int{2, 3})},
	})
	cpuSets := ShardCPUSets(shards)
	if len(cpuSets) != 2 {
		t.Fatalf("len(cpuSets) = %d, want 2", len(cpuSets))
	}
	if cpuSets[1][0] != 2 || cpuSets[1][1] != 3 {
		t.Fatalf("cpuSets[1] = %v, want [2 3]", cpuSets[1])
	}
}

func TestSelectShardByCPUSetUsesLargestOverlap(t *testing.T) {
	shards := NUMANodeShards([]NUMANode{
		{ID: 0, CPUs: NewCPUSet([]int{0, 1, 2, 3})},
		{ID: 1, CPUs: NewCPUSet([]int{4, 5, 6, 7})},
	})

	if got := SelectShardByCPUSet(shards, NewCPUSet([]int{4, 5}), 0); got != 1 {
		t.Fatalf("SelectShardByCPUSet() = %d, want 1", got)
	}
}
