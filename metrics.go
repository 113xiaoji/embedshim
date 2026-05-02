package embedshim

import (
	"expvar"
	"sync"
	"sync/atomic"
)

var (
	monitorMetricsOnce sync.Once
	monitorMetrics     atomic.Value
)

func publishMonitorMetrics(m *monitor) {
	monitorMetrics.Store(m)
	monitorMetricsOnce.Do(func() {
		expvar.Publish("embedshim_pidfd", expvar.Func(func() interface{} {
			current := monitorMetrics.Load()
			if current == nil {
				return nil
			}
			return current.(*monitor).metricsSnapshot()
		}))
	})
}
