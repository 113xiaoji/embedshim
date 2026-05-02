package topology

type Shard struct {
	ID   int
	CPUs CPUSet
}

func NUMANodeShards(nodes []NUMANode) []Shard {
	shards := make([]Shard, 0, len(nodes))
	for _, node := range nodes {
		if node.CPUs.Len() == 0 {
			continue
		}
		shards = append(shards, Shard{
			ID:   node.ID,
			CPUs: node.CPUs,
		})
	}
	return shards
}

func ShardCPUSets(shards []Shard) [][]int {
	out := make([][]int, 0, len(shards))
	for _, shard := range shards {
		if shard.CPUs.Len() == 0 {
			continue
		}
		out = append(out, shard.CPUs.List())
	}
	return out
}

func SelectShardByCPUSet(shards []Shard, cpus CPUSet, traceID uint64) int {
	if len(shards) == 0 {
		return -1
	}

	bestIdx := -1
	bestOverlap := 0
	target := cpus.List()
	for idx, shard := range shards {
		overlap := countOverlap(target, shard.CPUs.List())
		if overlap > bestOverlap {
			bestIdx = idx
			bestOverlap = overlap
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	return int(traceID % uint64(len(shards)))
}

func countOverlap(a, b []int) int {
	count := 0
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			count++
			i++
			j++
		case a[i] < b[j]:
			i++
		default:
			j++
		}
	}
	return count
}
