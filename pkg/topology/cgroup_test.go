package topology

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadCgroupPlacementReadsEffectiveSets(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "cpuset.cpus.effective"), []byte("0-3\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cpuset.mems.effective"), []byte("0,2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	placement, err := LoadCgroupPlacement(root)
	if err != nil {
		t.Fatalf("LoadCgroupPlacement() error = %v", err)
	}
	if got, want := placement.EffectiveCPUs.List(), []int{0, 1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("EffectiveCPUs = %v, want %v", got, want)
	}
	if got, want := placement.EffectiveMems.List(), []int{0, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("EffectiveMems = %v, want %v", got, want)
	}
}
