package topology

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type CPUSet struct {
	cpus []int
}

func ParseCPUSet(input string) (CPUSet, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return CPUSet{}, nil
	}

	var cpus []int
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return CPUSet{}, fmt.Errorf("empty cpu list segment in %q", input)
		}

		if strings.Contains(part, "-") {
			bounds := strings.Split(part, "-")
			if len(bounds) != 2 {
				return CPUSet{}, fmt.Errorf("invalid cpu range %q", part)
			}
			start, err := parseCPUIndex(bounds[0])
			if err != nil {
				return CPUSet{}, err
			}
			end, err := parseCPUIndex(bounds[1])
			if err != nil {
				return CPUSet{}, err
			}
			if start > end {
				return CPUSet{}, fmt.Errorf("invalid descending cpu range %q", part)
			}
			for cpu := start; cpu <= end; cpu++ {
				cpus = append(cpus, cpu)
			}
			continue
		}

		cpu, err := parseCPUIndex(part)
		if err != nil {
			return CPUSet{}, err
		}
		cpus = append(cpus, cpu)
	}

	return NewCPUSet(cpus), nil
}

func NewCPUSet(cpus []int) CPUSet {
	if len(cpus) == 0 {
		return CPUSet{}
	}

	cpus = append([]int(nil), cpus...)
	sort.Ints(cpus)
	out := make([]int, 0, len(cpus))
	for _, cpu := range cpus {
		if len(out) == 0 || out[len(out)-1] != cpu {
			out = append(out, cpu)
		}
	}
	return CPUSet{cpus: out}
}

func (s CPUSet) List() []int {
	out := make([]int, len(s.cpus))
	copy(out, s.cpus)
	return out
}

func (s CPUSet) Len() int {
	return len(s.cpus)
}

func (s CPUSet) String() string {
	if len(s.cpus) == 0 {
		return ""
	}

	var parts []string
	start := s.cpus[0]
	prev := s.cpus[0]
	for _, cpu := range s.cpus[1:] {
		if cpu == prev+1 {
			prev = cpu
			continue
		}
		parts = append(parts, formatCPURange(start, prev))
		start = cpu
		prev = cpu
	}
	parts = append(parts, formatCPURange(start, prev))
	return strings.Join(parts, ",")
}

func parseCPUIndex(input string) (int, error) {
	cpu, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return 0, fmt.Errorf("invalid cpu index %q: %w", input, err)
	}
	if cpu < 0 {
		return 0, fmt.Errorf("invalid negative cpu index %d", cpu)
	}
	return cpu, nil
}

func formatCPURange(start, end int) string {
	if start == end {
		return strconv.Itoa(start)
	}
	return fmt.Sprintf("%d-%d", start, end)
}
