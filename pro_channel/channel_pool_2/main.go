package main

import (
	"fmt"
	"sync"
	"time"
)

type Job func()

type worker struct {
	workerPool chan *worker
	jobChannel chan Job
	stop       chan struct{}
}

func newWorker(pool chan *worker) *worker {
	return &worker{
		workerPool: pool,
		jobChannel: make(chan Job),
		stop:       make(chan struct{}),
	}
}

func (w *worker) start() {
	go func() {
		for {
			w.workerPool <- w
			select {
			case job := <-w.jobChannel:
				job()
			case <-w.stop:
				w.stop <- struct{}{}
				return
			}
		}
	}()
}

type dispatcher struct {
	workerPool chan *worker
	jobQueue   chan Job
	stop       chan struct{}
}

func (d *dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			w := <-d.workerPool
			w.jobChannel <- job
		case <-d.stop:
			for i := 0; i < cap(d.workerPool); i++ {
				w := <-d.workerPool
				w.stop <- struct{}{}
				<-w.stop
			}
			d.stop <- struct{}{}
			return
		}
	}
}

func newDispatcher(workerPool chan *worker, jobQueue chan Job) *dispatcher {
	d := &dispatcher{
		workerPool: workerPool,
		jobQueue:   jobQueue,
		stop:       make(chan struct{}),
	}

	for i := 0; i < cap(d.workerPool); i++ {
		worker := newWorker(d.workerPool)
		worker.start()
	}
	go d.dispatch()
	return d
}

type Pool struct {
	JobQueue   chan Job
	dispatcher *dispatcher
	wg         sync.WaitGroup
}

func NewPool(numWorkers int, jobQueueLen int) *Pool {
	jobQueue := make(chan Job, jobQueueLen)
	workerPool := make(chan *worker, numWorkers)

	return &Pool{
		JobQueue:   jobQueue,
		dispatcher: newDispatcher(workerPool, jobQueue),
	}
}

func (p *Pool) JobDone() {
	p.wg.Done()
}

func (p *Pool) WaitCount(count int) {
	p.wg.Add(count)
}

func (p *Pool) WaitAll() {
	p.wg.Wait()
}

func (p *Pool) Release() {
	p.dispatcher.stop <- struct{}{}
	<-p.dispatcher.stop
}

func main() {
	start := time.Now()

	//numCPUs := runtime.NumCPU()
	//runtime.GOMAXPROCS(numCPUs)
	//
	//pool := NewPool(4, 10)
	//defer pool.Release()
	//for i := 0; i < 1000000; i++ {
	//	count := i
	//	pool.JobQueue <- func() {
	//		fmt.Printf("I am worker! Number %d\n", count)
	//	}
	//}
	//time.Sleep(1 * time.Second)
	//

	for i := 0; i < 1000000; i++ {
		fmt.Printf("I am worker! Number %d\n", i)
	}

	terminal := time.Since(start)
	fmt.Printf("耗时：%s \r\n", terminal)
}
