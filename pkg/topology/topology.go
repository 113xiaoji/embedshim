package topology

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type NUMANode struct {
	ID   int
	CPUs CPUSet
}

func LoadNUMANodes(sysfsRoot string) ([]NUMANode, error) {
	nodeRoot := filepath.Join(sysfsRoot, "devices", "system", "node")
	entries, err := os.ReadDir(nodeRoot)
	if err != nil {
		return nil, err
	}

	var nodes []NUMANode
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "node") {
			continue
		}

		id, err := strconv.Atoi(strings.TrimPrefix(entry.Name(), "node"))
		if err != nil {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(nodeRoot, entry.Name(), "cpulist"))
		if err != nil {
			return nil, fmt.Errorf("read node%d cpulist: %w", id, err)
		}

		cpus, err := ParseCPUSet(string(raw))
		if err != nil {
			return nil, fmt.Errorf("parse node%d cpulist: %w", id, err)
		}
		nodes = append(nodes, NUMANode{
			ID:   id,
			CPUs: cpus,
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
	return nodes, nil
}
