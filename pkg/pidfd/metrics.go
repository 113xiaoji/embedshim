package pidfd

import "sync/atomic"

// Metrics tracks pidfd epoll and callback-dispatch health.
type Metrics struct {
	epollEvents        uint64
	callbacksQueued    uint64
	callbacksStarted   uint64
	callbacksCompleted uint64
	callbackErrors     uint64
	dispatchErrors     uint64
	queueBackpressure  uint64
	affinityErrors     uint64
}

// MetricsSnapshot is a point-in-time copy of pidfd monitor counters.
type MetricsSnapshot struct {
	EpollEvents        uint64
	CallbacksQueued    uint64
	CallbacksStarted   uint64
	CallbacksCompleted uint64
	CallbackErrors     uint64
	DispatchErrors     uint64
	QueueBackpressure  uint64
	AffinityErrors     uint64
	CallbackQueueLen   int
	CallbackQueueCap   int
}

func (m *Metrics) snapshot(queueLen, queueCap int) MetricsSnapshot {
	if m == nil {
		return MetricsSnapshot{
			CallbackQueueLen: queueLen,
			CallbackQueueCap: queueCap,
		}
	}
	return MetricsSnapshot{
		EpollEvents:        atomic.LoadUint64(&m.epollEvents),
		CallbacksQueued:    atomic.LoadUint64(&m.callbacksQueued),
		CallbacksStarted:   atomic.LoadUint64(&m.callbacksStarted),
		CallbacksCompleted: atomic.LoadUint64(&m.callbacksCompleted),
		CallbackErrors:     atomic.LoadUint64(&m.callbackErrors),
		DispatchErrors:     atomic.LoadUint64(&m.dispatchErrors),
		QueueBackpressure:  atomic.LoadUint64(&m.queueBackpressure),
		AffinityErrors:     atomic.LoadUint64(&m.affinityErrors),
		CallbackQueueLen:   queueLen,
		CallbackQueueCap:   queueCap,
	}
}

func (m *Metrics) incEpollEvents() {
	if m != nil {
		atomic.AddUint64(&m.epollEvents, 1)
	}
}

func (m *Metrics) incCallbacksQueued() {
	if m != nil {
		atomic.AddUint64(&m.callbacksQueued, 1)
	}
}

func (m *Metrics) incCallbacksStarted() {
	if m != nil {
		atomic.AddUint64(&m.callbacksStarted, 1)
	}
}

func (m *Metrics) incCallbacksCompleted() {
	if m != nil {
		atomic.AddUint64(&m.callbacksCompleted, 1)
	}
}

func (m *Metrics) incCallbackErrors() {
	if m != nil {
		atomic.AddUint64(&m.callbackErrors, 1)
	}
}

func (m *Metrics) incDispatchErrors() {
	if m != nil {
		atomic.AddUint64(&m.dispatchErrors, 1)
	}
}

func (m *Metrics) incQueueBackpressure() {
	if m != nil {
		atomic.AddUint64(&m.queueBackpressure, 1)
	}
}

func (m *Metrics) incAffinityErrors() {
	if m != nil {
		atomic.AddUint64(&m.affinityErrors, 1)
	}
}
