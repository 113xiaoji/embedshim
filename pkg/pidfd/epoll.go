package pidfd

import (
	"fmt"
	"sync"

	"golang.org/x/sys/unix"
)

type pidOnClose func() error

// Epoller is used to monitor PID file descriptors.
//
// When the process that PID file descriptor refers to terminates, these
// interfaces indicate the file descriptor as readable. Then Epoller will close
// the PID file descriptor and call the callback.
type Epoller struct {
	mu         sync.Mutex
	efd        int
	closeOnce  sync.Once
	maxEvents  int
	dispatcher *callbackDispatcher
	metrics    *Metrics

	fdOnCloses map[FD]pidOnClose
}

func NewEpoller(opts ...Option) (*Epoller, error) {
	options, err := newOptions(opts...)
	if err != nil {
		return nil, err
	}

	efd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}

	dispatcher := newCallbackDispatcherWithCPUSets(
		options.callbackWorkers,
		options.callbackQueueDepth,
		options.metrics,
		options.workerCPUSets,
	)
	dispatcher.start()

	e := &Epoller{
		efd:        efd,
		maxEvents:  options.maxEvents,
		dispatcher: dispatcher,
		metrics:    options.metrics,
		fdOnCloses: make(map[FD]pidOnClose),
	}
	return e, nil
}

// Add monitors the PID file descriptor and registers the onClose for it.
func (e *Epoller) Add(fd FD, onClose func() error) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.fdOnCloses[fd]; ok {
		return fmt.Errorf("the pidfd %v is exist", fd)
	}

	ev := unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(fd),
	}

	if err := unix.EpollCtl(e.efd, unix.EPOLL_CTL_ADD, int(fd), &ev); err != nil {
		return fmt.Errorf("failed to monitor pidfd: %w", err)
	}

	e.fdOnCloses[fd] = onClose
	return nil
}

// Run starts to monitor the event on PID file descriptor.
func (e *Epoller) Run() error {
	events := make([]unix.EpollEvent, e.maxEvents)

	for {
		n, err := unix.EpollWait(e.efd, events, -1)
		if err != nil {
			// EINTR: The call was interrupted by a signal handler
			// before either any of the requested events occurred
			// or the timeout expired
			if err == unix.EINTR {
				continue
			}
			return fmt.Errorf("failed to wait pidfd events: %w", err)
		}

		for i := 0; i < n; i++ {
			e.metrics.incEpollEvents()

			fd := FD(events[i].Fd)

			err := unix.EpollCtl(e.efd, unix.EPOLL_CTL_DEL, int(fd), &unix.EpollEvent{})
			if err != nil {
				return fmt.Errorf("failed to remove pidfd from interest list: %w", err)
			}

			e.mu.Lock()

			onClose := e.fdOnCloses[fd]
			delete(e.fdOnCloses, fd)

			e.mu.Unlock()

			unix.Close(int(fd))
			if onClose == nil {
				continue
			}
			if err := e.dispatcher.dispatch(onClose); err != nil {
				return err
			}
		}
	}
}

// Metrics returns a point-in-time snapshot of epoll and callback counters.
func (e *Epoller) Metrics() MetricsSnapshot {
	return e.dispatcher.metricsSnapshot()
}

// Close stops the monitor.
func (e *Epoller) Close() error {
	e.closeOnce.Do(func() {
		unix.Close(e.efd)
		e.dispatcher.close()
	})
	return nil
}
