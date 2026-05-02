package embedshim

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPIDFDWorkerCPUSetsFromConfigDisabledByDefault(t *testing.T) {
	if got := pidfdWorkerCPUSetsFromConfig(&Config{}); got != nil {
		t.Fatalf("pidfdWorkerCPUSetsFromConfig() = %v, want nil", got)
	}
}

func TestPIDFDWorkerCPUSetsFromConfigLoadsNUMATopology(t *testing.T) {
	root := t.TempDir()
	nodeRoot := filepath.Join(root, "devices", "system", "node")
	for _, node := range []struct {
		name    string
		cpulist string
	}{
		{name: "node0", cpulist: "0-1\n"},
		{name: "node1", cpulist: "2-3\n"},
	} {
		if err := os.MkdirAll(filepath.Join(nodeRoot, node.name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(nodeRoot, node.name, "cpulist"), []byte(node.cpulist), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got := pidfdWorkerCPUSetsFromConfig(&Config{
		ARMAffinityMode:        armAffinityModeRuntimeOnly,
		ARMAffinitySharding:    armAffinityShardingNUMA,
		ARMAffinityBindWorkers: true,
		TopologySysfsRoot:      root,
	})
	if len(got) != 2 {
		t.Fatalf("len(cpuSets) = %d, want 2", len(got))
	}
	if got[1][0] != 2 || got[1][1] != 3 {
		t.Fatalf("cpuSets[1] = %v, want [2 3]", got[1])
	}
}
