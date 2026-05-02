package embedshim

import (
	"github.com/fuweid/embedshim/pkg/pidfd"
)

type exitMonitor interface {
	traceInitProcess(*initProcess) error
	repollingInitProcess(*initProcess) error
	cleanInitProcessTraceEvent(*initProcess) error
	traceExecProcess(*execProcess, uint32) (pidfd.FD, error)
	recordExecExitedStatus(*execProcess, uint32, bool, uint32) error
	pollExecProcess(*execProcess, pidfd.FD) error
	metricsSnapshot() pidfd.MetricsSnapshot
}

var _ exitMonitor = (*monitor)(nil)
