package signals

import (
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"syscall"
	"testing"
)

func TestDispatchChannel(t *testing.T) {
	assert := assert.New(t)
	ch := make(chan os.Signal)
	wg := sync.WaitGroup{}

	s := DispatchChannel(ch)
	assert.NotNil(s)

	// No handlers
	s.Emit(os.Interrupt)

	// One handler
	cntTerm := 0
	wg.Add(1)
	s.On(syscall.SIGTERM, func() { cntTerm++; wg.Done() })
	s.Emit(os.Interrupt)
	assert.Equal(0, cntTerm)
	s.Emit(syscall.SIGTERM)
	wg.Wait()
	assert.Equal(1, cntTerm)

	// More handlers
	cntHup := 0
	wg.Add(1)
	s.OnHUP(func() { cntHup++; wg.Done() })
	s.Emit(syscall.SIGHUP)
	wg.Wait()
	assert.Equal(1, cntTerm)
	assert.Equal(1, cntHup)
	wg.Add(1)
	s.Emit(syscall.SIGTERM)
	wg.Wait()
	assert.Equal(2, cntTerm)

	// More handlers
	cntUsr1 := 0
	wg.Add(1)
	s.OnUSR1(func() { cntUsr1++; wg.Done() })
	s.Emit(syscall.SIGUSR1)
	wg.Wait()
	assert.Equal(2, cntTerm)
	assert.Equal(1, cntHup)
	assert.Equal(1, cntUsr1)
	wg.Add(1)
	s.Emit(syscall.SIGTERM)
	wg.Wait()
	assert.Equal(3, cntTerm)

	// More handlers
	cntUsr2 := 0
	wg.Add(1)
	s.OnUSR2(func() { cntUsr2++; wg.Done() })
	s.Emit(syscall.SIGUSR2)
	wg.Wait()
	assert.Equal(3, cntTerm)
	assert.Equal(1, cntHup)
	assert.Equal(1, cntUsr1)
	assert.Equal(1, cntUsr2)
	wg.Add(1)
	s.Emit(syscall.SIGUSR1)
	wg.Wait()
	assert.Equal(2, cntUsr1)
	assert.Equal(1, cntUsr2)

	// More handlers
	cntShutdown := 0
	wg.Add(2)
	s.OnShutdown(func() { cntShutdown++; wg.Done() })
	s.Shutdown()
	wg.Wait()
	assert.Equal(4, cntTerm, "Shutdown() must send syscall.SIGTERM")
	assert.Equal(1, cntShutdown)

	// Components
	cnt := new(counter)
	s.RegisterCloser(cnt)
	s.RegisterReloader(cnt)
	wg.Add(2)
	s.Shutdown()
	wg.Wait()
	assert.Equal(5, cntTerm)
	assert.Equal(1, cnt.closes)
	assert.Equal(0, cnt.reloads)
	wg.Add(1)
	s.Emit(syscall.SIGHUP)
	wg.Wait()
	assert.Equal(2, cntHup)
	assert.Equal(1, cnt.closes)
	assert.Equal(1, cnt.reloads)

	// Nil ok
	s.RegisterCloser(nil)
	s.RegisterReloader(nil)
	wg.Add(3)
	s.Emit(syscall.SIGHUP)
	s.Emit(syscall.SIGTERM)
	wg.Wait()
}

type counter struct {
	closes  int
	reloads int
}

func (c *counter) Close() error {
	c.closes++
	return nil
}

func (c *counter) Reload() error {
	c.reloads++
	return nil
}
