package main

import (
	"sync"
)

/***********************线程安全的优先级加减*****************************/

type safepending struct {
	pending int
	*sync.RWMutex
}

func (s *safepending) Inc() {
	s.Lock()
	s.pending++
	s.Unlock()
}

func (s *safepending) Dec() {
	s.Lock()
	s.pending--
	s.Unlock()
}

func (s *safepending) Get() int {
	s.RLock()
	n := s.pending
	s.RUnlock()
	return n
}

/*
*************小顶堆的工作池****************
*/
type Request struct {
	fn   func() int
	data []byte
	op   int
	c    chan int
}

type Worker struct {
	req     chan Request
	pending int
	index   int
	done    chan struct{}
}

type Pool []*Worker

func (p Pool) Len() int {
	return len(p)
}

func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending
}

func (p Pool) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].index = i
	p[j].index = j
}

func (p *Pool) Push(x interface{}) {
	n := len(*p)
	item := x.(*Worker)
	item.index = n
	*p = append(*p, item)
}

func (p *Pool) Pop() interface{} {
	old := *p
	n := len(*p)
	item := old[n-1]
	*p = old[:n-1]
	return item
}

/******************动态增删任务实现（独立线程访问Map，使Map变的安全）*******************/
type job struct {
}

type jobPair struct {
	key   string
	value *job
}

type worker struct {
	jobqueue map[string]*job
	jobadd   chan *jobPair
	jobdel   chan string
	pending  safepending
}

func (w *worker) Run() {

}

func (w *worker) PushJob(user string, job *job) {
	pair := &jobPair{
		key:   user,
		value: job,
	}
	w.jobadd <- pair
}

func (w *worker) RemoveJob(user string) {
	w.jobdel <- user
}

func (w *worker) insertJob(key string, value *job) error {
	w.jobqueue[key] = value
	w.pending.Inc()
	return nil
}

func (w *worker) deleteJob(key string) {
	delete(w.jobqueue, key)
	w.pending.Dec()
}

/*****************线程池实现*****************/
var (
	MaxWorks = 10000
	MaxQueue = 1000 //通道缓存队列数
)

func (w *Worker) Run() {
	go func() {
		for {
			select {
			case req := <-w.req:
				req.c <- req.fn()
			case <-w.done:
				break
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.done <- struct{}{}
}

func main() {

	//fmt.Println("初始化工作池")
	//pool := new(Pool)
	//for i := 0; i < 15; i++ {
	//	work := &Worker{
	//		req:      make(chan Request, MaxQueue),
	//		pengding: rand.Intn(100),
	//		index:    i,
	//	}
	//	fmt.Println("pengding:", work.pengding, "--- i:", i)
	//	heap.Push(pool, work)
	//}
	//heap.Init(pool)
	//
	//fmt.Println("初始化工作池-成功：", pool)

	//for pool.Len() > 0 {
	//	worker := heap.Pop(pool).(*Worker)
	//	fmt.Println("pengding:", worker.pending, "--- i:", worker.index)
	//}

	//worker1 := heap.Pop(pool).(*Worker)
	//
	//fmt.Println("priority:", worker1.priority, "--- i:", worker1.index)

}
