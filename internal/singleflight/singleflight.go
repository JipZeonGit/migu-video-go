package singleflight

import (
	"fmt"
	"sync"
)

// call represents an in-flight or completed singleflight.Do call.
type call struct {
	wg  sync.WaitGroup
	val string
	err error
}

// Group ensures that only one execution of a given key is in-flight at a time.
// Duplicate callers wait for the original to complete and receive the same result.
type Group struct {
	mu    sync.Mutex
	calls map[string]*call
}

// Do executes and returns the results of the given function, making sure that
// only one execution is in-flight for a given key at a time. If a duplicate
// call comes in, the duplicate caller waits for the original to complete and
// receives the same results.
func (g *Group) Do(key string, fn func() (string, error)) (string, error) {
	g.mu.Lock()
	if g.calls == nil {
		g.calls = make(map[string]*call)
	}
	if c, ok := g.calls[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()

	g.call(c, fn)

	return c.val, c.err
}

// call executes fn and ensures wg.Done is always called (even on panic).
func (g *Group) call(c *call, fn func() (string, error)) {
	defer func() {
		if r := recover(); r != nil {
			c.err = fmt.Errorf("singleflight panic: %v", r)
		}
		c.wg.Done()
		g.mu.Lock()
		// 清理：从 map 中删除此 key（通过遍历找到并删除）
		for k, v := range g.calls {
			if v == c {
				delete(g.calls, k)
				break
			}
		}
		g.mu.Unlock()
	}()

	c.val, c.err = fn()
}
