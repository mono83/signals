package signals

import (
	"github.com/mono83/slf"
	"github.com/mono83/slf/wd"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Signals
const (
	SHUTDOWN = syscall.SIGTERM
	RELOAD   = syscall.SIGHUP

	SIGINT  = syscall.SIGINT
	SIGTERM = syscall.SIGTERM
	SIGHUP  = syscall.SIGHUP
	SIGUSR1 = syscall.SIGUSR1
	SIGUSR2 = syscall.SIGUSR2
)

// NewDefaultDispatcher returns a dispatcher for SIGINT, SIGTERM, SIGHUP, SIGUSR1, SIGUSR2
func NewDefaultDispatcher() ExtendedDispatcher {
	return DispatchSignals(SIGINT, SIGTERM, SIGHUP, SIGUSR1, SIGUSR2)
}

// DispatchSignals registers listener for required signals and returns
// a dispatcher for that events
func DispatchSignals(signals ...os.Signal) ExtendedDispatcher {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)

	return DispatchChannel(c)
}

// DispatchChannel returns dispatcher, built on top of provided channel
func DispatchChannel(src chan os.Signal) ExtendedDispatcher {
	s := new(signals)

	// Injecting channel
	s.source = src

	// Initializing logger
	s.log = wd.NewLogger("signals")

	// Building handlers map
	s.handlers = map[os.Signal][]func(){}

	go s.handle()

	return s
}

type signals struct {
	m        sync.Mutex
	handlers map[os.Signal][]func()

	source chan os.Signal
	log    slf.Logger
}

func (s *signals) handle() {
	// Listening
	for sig := range s.source {
		// Reading registered handlers for that signal
		s.m.Lock()
		handlers, _ := s.handlers[sig]
		s.m.Unlock()

		s.log.Info("Received signal :name to process by :count handlers", wd.NameParam(sig.String()), wd.CountParam(len(handlers)))

		if len(handlers) > 0 {
			go func(hs []func()) {
				for _, handler := range hs {
					handler()
				}
			}(handlers)
		}
	}
}

func (s *signals) Shutdown() {
	s.Emit(SHUTDOWN)
}

func (s *signals) Reload() {
	s.Emit(RELOAD)
}

func (s *signals) On(sig os.Signal, target func()) {
	s.m.Lock()
	defer s.m.Unlock()

	s.handlers[sig] = append(s.handlers[sig], target)
	s.log.Debug("Registered handler (:count total) for :name", wd.NameParam(sig.String()), wd.CountParam(len(s.handlers)))
}

func (s *signals) Emit(sig os.Signal) {
	if s != nil {
		s.source <- sig
	}
}

func (s *signals) OnShutdown(target func()) {
	s.On(SIGINT, target)
	s.On(SIGTERM, target)
}

func (s *signals) OnHUP(target func()) {
	s.On(SIGHUP, target)
}

func (s *signals) OnUSR1(target func()) {
	s.On(SIGUSR1, target)
}

func (s *signals) OnUSR2(target func()) {
	s.On(SIGUSR2, target)
}

func (s *signals) RegisterCloser(c Closer) {
	if c != nil {
		s.OnShutdown(func() {
			err := c.Close()
			if err != nil {
				s.log.Error("Closer on shutdown failed with :err", wd.ErrParam(err))
			}
		})
	}
}

func (s *signals) RegisterReloader(r Reloader) {
	if r != nil {
		s.OnHUP(func() {
			err := r.Reload()
			if err != nil {
				s.log.Error("Reloader failed with :err", wd.ErrParam(err))
			}
		})
	}
}
