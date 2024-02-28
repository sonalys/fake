package boilerplate

import (
	"sync"
	"testing"
)

type (
	Config interface {
		Repeat(times int)
	}

	call[T any] struct {
		lock   *sync.Mutex
		repeat int
		cur    int
		hooks  []T
	}

	mock[T any] struct {
		lock  *sync.Mutex
		calls []*call[T]
	}
)

var (
	RepeatForever int = -1
)

func setupLocker() *sync.Mutex { return &sync.Mutex{} }

func newMock[T any](t *testing.T) mock[T] {
	value := mock[T]{
		lock: &sync.Mutex{},
	}
	t.Cleanup(func() {
		value.AssertExpectations(t)
	})
	return value
}

func (c *mock[T]) AssertExpectations(t *testing.T) bool {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()

	var missingCalls int
	for _, call := range c.calls {
		call.lock.Lock()
		if call.repeat <= 0 {
			call.repeat = 1
		}
		missingCalls += call.repeat*len(call.hooks) - call.cur
		call.lock.Unlock()
	}
	if missingCalls > 0 {
		t.Errorf("%d more calls expected to func %T", missingCalls, c.calls[0].hooks[0])
		return false
	}
	return true
}

func (c *call[T]) Repeat(times int) {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()
	c.repeat = times
}

func (c *call[T]) draw() (f T, empty bool) {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()
	f = c.hooks[c.cur]
	c.cur++
	if c.cur == len(c.hooks) {
		c.repeat--
		c.cur = c.cur % len(c.hooks)
	}
	return f, c.repeat == -1
}

func (c *mock[T]) call() (*T, bool) {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(c.calls) == 0 {
		return nil, false
	}
	first, empty := c.calls[0].draw()
	if empty {
		c.calls = c.calls[1:]
	}
	return &first, true
}

func (c *mock[T]) append(f ...T) Config {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()
	call := &call[T]{
		hooks: f,
		lock:  &sync.Mutex{},
	}
	c.calls = append(c.calls, call)
	return call
}
