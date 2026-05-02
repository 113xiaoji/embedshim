package topology

import (
	"reflect"
	"testing"
)

func TestParseCPUSetExpandsRangesAndSingles(t *testing.T) {
	set, err := ParseCPUSet("0-3,8,10-11")
	if err != nil {
		t.Fatalf("ParseCPUSet() error = %v", err)
	}
	want := []int{0, 1, 2, 3, 8, 10, 11}
	if got := set.List(); !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %v, want %v", got, want)
	}
	if got := set.String(); got != "0-3,8,10-11" {
		t.Fatalf("String() = %q, want %q", got, "0-3,8,10-11")
	}
}

func TestParseCPUSetRejectsDescendingRange(t *testing.T) {
	if _, err := ParseCPUSet("4-2"); err == nil {
		t.Fatal("ParseCPUSet() error = nil, want descending range error")
	}
}
