package singleflight

import "sync"

// call代表正在进行中的请求
type call struct {
	wg  sync.WaitGroup // 阻塞调用这个call的其他请求
	val interface{}    // 请求的结果
	err error
}

// Group是主体数据结构，管理key和对应的call
type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // 一个key对应一个call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // 阻塞，等待锁释放
		return c.val, c.err // 请求结束， 返回
	}
	c := new(call)
	c.wg.Add(1)  // 锁 + 1
	g.m[key] = c // 记录call, 表示key已经有call在处理
	g.mu.Unlock()

	c.val, c.err = fn() // 发起请求
	c.wg.Done()         // 锁 - 1

	g.mu.Lock()
	delete(g.m, key) // 更新map
	g.mu.Unlock()

	return c.val, c.err
}
