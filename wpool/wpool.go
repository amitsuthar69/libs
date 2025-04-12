// worker pool
package wpool

/*
- shutdown
- dynamc scaling
*/

import "sync"

type work func() any

type WPool struct {
	pool chan work
	res  chan any
	wg   sync.WaitGroup
}

func NewWPool(workers int) *WPool {
	wpool := &WPool{
		pool: make(chan work),
		res:  make(chan any),
	}

	wpool.wg.Add(workers)
	for range workers {
		go func() {
			defer wpool.wg.Done()
			for work := range wpool.pool {
				wpool.res <- work()
			}
		}()
	}
	return wpool
}

func (wp *WPool) AddWork(work work) {
	wp.pool <- work
}

func (wp *WPool) Result() <-chan any {
	return wp.res
}

func (wp *WPool) Close() {
	close(wp.pool)
	wp.wg.Wait()
}
