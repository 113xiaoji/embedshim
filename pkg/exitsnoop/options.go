package exitsnoop

import (
	"fmt"

	"github.com/cilium/ebpf"
)

const (
	defaultMapMaxEntries uint32 = 2048
	minMapMaxEntries     uint32 = 16
)

type options struct {
	mapMaxEntries uint32
}

type Option func(*options) error

func newOptions(opts ...Option) (*options, error) {
	out := &options{
		mapMaxEntries: defaultMapMaxEntries,
	}
	for _, opt := range opts {
		if err := opt(out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func WithMapMaxEntries(entries uint32) Option {
	return func(o *options) error {
		if entries == 0 {
			return nil
		}
		if entries < minMapMaxEntries {
			return fmt.Errorf("bpf map max entries %d is smaller than minimum %d", entries, minMapMaxEntries)
		}
		o.mapMaxEntries = entries
		return nil
	}
}

func applyMapSizing(spec *ebpf.CollectionSpec, entries uint32) error {
	for _, name := range []string{bpfMapTracingTasks, bpfMapExitedEvents} {
		m, ok := spec.Maps[name]
		if !ok {
			return fmt.Errorf("bpf map %s is missing from collection spec", name)
		}
		m.MaxEntries = entries
	}
	return nil
}

func validateMapMaxEntries(name string, got, want uint32) error {
	if got == want {
		return nil
	}
	return fmt.Errorf("pinned bpf map %s has max entries %d, want %d; remove the pinned exitsnoop objects or keep the previous capacity", name, got, want)
}
