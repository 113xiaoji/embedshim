package pidfd

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

type callbackDispatcher struct {
	queue chan pidOnClose
	done  chan struct{}
	once  sync.Once
	wg    sync.WaitGroup

	workers int
	closed  uint32
	metrics *Metrics
	cpuSets [][]int
}

func newCallbackDispatcher(workers int, queueDepth int, metrics ...*Metrics) *callbackDispatcher {
	var m *Metrics
	if len(metrics) > 0 {
		m = metrics[0]
	}
	if m == nil {
		m = &Metrics{}
	}

	return &callbackDispatcher{
		queue:   make(chan pidOnClose, queueDepth),
		done:    make(chan struct{}),
		workers: workers,
		metrics: m,
	}
}

func newCallbackDispatcherWithCPUSets(workers int, queueDepth int, metrics *Metrics, cpuSets [][]int) *callbackDispatcher {
	dispatcher := newCallbackDispatcher(workers, queueDepth, metrics)
	dispatcher.cpuSets = copyCPUSets(cpuSets)
	return dispatcher
}

func (d *callbackDispatcher) start() {
	for i := 0; i < d.workers; i++ {
		cpuSet := d.workerCPUSet(i)
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			if len(cpuSet) > 0 {
				runtime.LockOSThread()
				defer runtime.UnlockOSThread()
				if err := setCurrentThreadAffinity(cpuSet); err != nil {
					d.metrics.incAffinityErrors()
				}
			}
			for {
				select {
				case callback := <-d.queue:
					if callback != nil {
						d.metrics.incCallbacksStarted()
						if err := callback(); err != nil {
							d.metrics.incCallbackErrors()
						}
						d.metrics.incCallbacksCompleted()
					}
				case <-d.done:
					return
				}
			}
		}()
	}
}

func (d *callbackDispatcher) dispatch(callback pidOnClose) error {
	if atomic.LoadUint32(&d.closed) != 0 {
		d.metrics.incDispatchErrors()
		return fmt.Errorf("pidfd callback dispatcher is closed")
	}

	select {
	case d.queue <- callback:
		d.metrics.incCallbacksQueued()
		return nil
	default:
		d.metrics.incQueueBackpressure()
	}

	select {
	case d.queue <- callback:
		d.metrics.incCallbacksQueued()
		return nil
	case <-d.done:
		d.metrics.incDispatchErrors()
		return fmt.Errorf("pidfd callback dispatcher is closed")
	}
}

func (d *callbackDispatcher) close() {
	d.once.Do(func() {
		atomic.StoreUint32(&d.closed, 1)
		close(d.done)
		d.wg.Wait()
	})
}

func (d *callbackDispatcher) metricsSnapshot() MetricsSnapshot {
	return d.metrics.snapshot(len(d.queue), cap(d.queue))
}

func (d *callbackDispatcher) workerCPUSet(workerID int) []int {
	if len(d.cpuSets) == 0 {
		return nil
	}
	return d.cpuSets[workerID%len(d.cpuSets)]
}

func copyCPUSets(cpuSets [][]int) [][]int {
	if len(cpuSets) == 0 {
		return nil
	}
	out := make([][]int, 0, len(cpuSets))
	for _, cpuSet := range cpuSets {
		if len(cpuSet) == 0 {
			continue
		}
		out = append(out, append([]int(nil), cpuSet...))
	}
	return out
}
