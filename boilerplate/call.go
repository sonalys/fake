package boilerplate

import (
	"sync"
	"testing"
)

type (
	// Config represents the configuration for all the functions passed to the On... function.
	Config interface {
		// Repeat sets how many times the function group should be called, note that if more than 1 function is given,
		// all the functions should repeat n times.
		// Repeat is 1 by default, meaning that functions can only be called 1 time.
		// Set Repeat(-1) to allow the group to repeat indefinitely.
		Repeat(times int)
		// Maybe sets the group as not required for AssertExpectations,
		// meaning that the function group will not fail the test if not called.
		Maybe()
	}

	Call[T any] struct {
		lock   *sync.Mutex
		repeat int
		maybe  bool
		cur    int
		hooks  []T
	}

	Mock[T any] struct {
		lock  *sync.Mutex
		calls []*Call[T]
	}
)

var (
	// RepeatForever, can be used with:
	//	OnMock().Repeat(mocks.RepeatForever).
	RepeatForever int = -1
)

// setupLocker is a sync.Once func to configure sync.Mutex in case newMock wasn't called.
func setupLocker() *sync.Mutex { return &sync.Mutex{} }

func (c *Call[T]) Repeat(times int) {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()
	c.repeat = times
}

func (c *Call[T]) Maybe() {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()
	c.maybe = true
}

func NewMock[T any](t *testing.T) Mock[T] {
	value := Mock[T]{
		lock: sync.OnceValue(setupLocker)(),
	}
	t.Cleanup(func() {
		value.AssertExpectations(t)
	})
	return value
}

// AssertExpectations asserts that all expected function calls have been called.
// Returns true if all expectations were met, otherwise returns false.
func (c *Mock[T]) AssertExpectations(t *testing.T) bool {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()
	var missingCalls int
	for _, call := range c.calls {
		call.lock.Lock()
		if call.maybe {
			continue
		}
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

// Draw is a card Draw design, in which each call determines when to get removed from the deck.
// if repeat == 0, the call gets removed from deck.
// setting repeat = -1 will skip this condition, allowing it to repeat indefinitely through the group.
func (c *Call[T]) Draw() (f T, empty bool) {
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

// Call returns a func of type T and a bool from the deck.
// It either returns (func, true) or (nil, false).
func (c *Mock[T]) Call() (*T, bool) {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(c.calls) == 0 {
		return nil, false
	}
	first, empty := c.calls[0].Draw()
	if empty {
		c.calls = c.calls[1:]
	}
	return &first, true
}

// Append creates a new card for the group of functions given, returning Config.
// With Config you can configure the group expectations.
func (c *Mock[T]) Append(f ...T) Config {
	c.lock = sync.OnceValue(setupLocker)()
	c.lock.Lock()
	defer c.lock.Unlock()
	call := &Call[T]{
		hooks: f,
		lock:  &sync.Mutex{},
	}
	c.calls = append(c.calls, call)
	return call
}
