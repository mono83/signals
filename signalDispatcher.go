package signals

import "os"

// SimpleDispatcher contains basic handler functionality
type SimpleDispatcher interface {
	// On method registers handler on custom signal
	On(os.Signal, func())
	// Emit sends provided signal and invokes corresponding
	// handlers
	Emit(os.Signal)
}

// Dispatcher describes structures, used to dispatch OS events
// to handler methods. Multiple handlers can be assigned to single event
type Dispatcher interface {
	SimpleDispatcher

	// OnShutdown registers handler for shutdown events
	OnShutdown(func())

	// OnHUP registers handler for reload (HUP) events
	OnHUP(func())

	// OnUSR1 registers handler for USR1 signal
	OnUSR1(func())

	// OnUSR2 registers handler for USR2 signal
	OnUSR2(func())

	// Shutdown method emits shutdown event
	Shutdown()
}

// ExtendedDispatcher describes dispatcher, that can provide own handlers
// for common usecases
type ExtendedDispatcher interface {
	Dispatcher

	// RegisterCloser registers handler on shutdown event, that will
	// invoke Close method on provided component
	RegisterCloser(Closer)

	// RegisterReloader registers handler on HUP (reload) event, that will
	// invoke Reload method on provided component
	RegisterReloader(Reloader)
}

// Closer is the interface that wraps the basic Close method.
type Closer interface {
	Close() error
}

// Reloader is the interface that wraps the basic Reload method.
type Reloader interface {
	Reload() error
}
