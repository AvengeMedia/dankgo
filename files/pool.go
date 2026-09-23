package files

import "sync"

// pool runs at most max workers and keeps none alive once the queue drains.
type pool struct {
	max int

	mu      sync.Mutex
	queue   []func()
	running int
	idle    *sync.Cond
}

func newPool(max int) *pool {
	if max < 1 {
		max = 1
	}
	p := &pool{max: max}
	p.idle = sync.NewCond(&p.mu)
	return p
}

func (p *pool) submit(task func()) {
	p.mu.Lock()
	p.queue = append(p.queue, task)
	if p.running >= p.max {
		p.mu.Unlock()
		return
	}
	p.running++
	p.mu.Unlock()
	go p.work()
}

func (p *pool) work() {
	for {
		p.mu.Lock()
		if len(p.queue) == 0 {
			p.running--
			p.idle.Broadcast()
			p.mu.Unlock()
			return
		}
		task := p.queue[0]
		p.queue = p.queue[1:]
		p.mu.Unlock()
		task()
	}
}

func (p *pool) drain() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for p.running > 0 || len(p.queue) > 0 {
		p.idle.Wait()
	}
}
