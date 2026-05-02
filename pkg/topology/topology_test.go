package topology

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadNUMANodesReadsAndSortsNodeCPULists(t *testing.T) {
	root := t.TempDir()
	nodeRoot := filepath.Join(root, "devices", "system", "node")
	for _, name := range []string{"node1", "node0", "not-a-node"} {
		if err := os.MkdirAll(filepath.Join(nodeRoot, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(nodeRoot, "node1", "cpulist"), []byte("4-7\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeRoot, "node0", "cpulist"), []byte("0-3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	nodes, err := LoadNUMANodes(root)
	if err != nil {
		t.Fatalf("LoadNUMANodes() error = %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("len(nodes) = %d, want 2", len(nodes))
	}
	if nodes[0].ID != 0 || nodes[1].ID != 1 {
		t.Fatalf("node IDs = %d,%d, want 0,1", nodes[0].ID, nodes[1].ID)
	}
	if got, want := nodes[1].CPUs.List(), []int{4, 5, 6, 7}; !reflect.DeepEqual(got, want) {
		t.Fatalf("node1 CPUs = %v, want %v", got, want)
	}
}
