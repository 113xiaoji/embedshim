package pidfd

import "fmt"

const (
	defaultMaxEvents          = 128
	defaultCallbackWorkers    = 1
	defaultCallbackQueueDepth = 128
)

type options struct {
	maxEvents          int
	callbackWorkers    int
	callbackQueueDepth int
	workerCPUSets      [][]int
	metrics            *Metrics
}

type Option func(*options) error

func newOptions(opts ...Option) (*options, error) {
	out := &options{
		maxEvents:          defaultMaxEvents,
		callbackWorkers:    defaultCallbackWorkers,
		callbackQueueDepth: defaultCallbackQueueDepth,
		metrics:            &Metrics{},
	}
	for _, opt := range opts {
		if err := opt(out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func WithMetrics(metrics *Metrics) Option {
	return func(o *options) error {
		if metrics != nil {
			o.metrics = metrics
		}
		return nil
	}
}

func WithWorkerCPUSets(cpuSets [][]int) Option {
	return func(o *options) error {
		if len(cpuSets) == 0 {
			return nil
		}
		out := make([][]int, 0, len(cpuSets))
		for _, cpuSet := range cpuSets {
			if len(cpuSet) == 0 {
				continue
			}
			copied := append([]int(nil), cpuSet...)
			out = append(out, copied)
		}
		o.workerCPUSets = out
		return nil
	}
}

func WithMaxEvents(maxEvents int) Option {
	return func(o *options) error {
		if maxEvents == 0 {
			return nil
		}
		if maxEvents < 0 {
			return fmt.Errorf("pidfd epoll max events %d must be greater than 0", maxEvents)
		}
		o.maxEvents = maxEvents
		return nil
	}
}

func WithCallbackWorkers(workers int) Option {
	return func(o *options) error {
		if workers == 0 {
			return nil
		}
		if workers < 0 {
			return fmt.Errorf("pidfd callback workers %d must be greater than 0", workers)
		}
		o.callbackWorkers = workers
		return nil
	}
}

func WithCallbackQueueDepth(depth int) Option {
	return func(o *options) error {
		if depth == 0 {
			return nil
		}
		if depth < 0 {
			return fmt.Errorf("pidfd callback queue depth %d must be greater than 0", depth)
		}
		o.callbackQueueDepth = depth
		return nil
	}
}
