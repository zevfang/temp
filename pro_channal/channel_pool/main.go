package main

import (
	"sync"
	"fmt"
	"time"
	"container/heap"
)

/***********************线程安全的优先级加减*****************************/

type safepending struct {
	pending int
	sync.RWMutex
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
*************小顶堆算法的工作池****************
*/

type Pool []*Worker

func (p Pool) Len() int {
	return len(p)
}

func (p Pool) Less(i, j int) bool {
	return p[i].pending.Get() < p[j].pending.Get()
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

/****************** 线程池实现 *******************/
type Job struct {
	fn     func() int
	result chan string
}

//type DataType struct {
//	v string
//}

type Worker struct {
	jobqueue  map[string]*Job
	broadcast chan string
	jobadd    chan *jobPair
	jobdel    chan string
	pending   safepending
	index     int
	done      chan struct{}
}

type jobPair struct {
	key   string
	value *Job
}

func NewWorker(idx, queue_limit, source_limit, jobreq_limit int) *Worker {
	return &Worker{
		jobqueue:  make(map[string]*Job, queue_limit),
		broadcast: make(chan string, source_limit),
		jobadd:    make(chan *jobPair, jobreq_limit),
		jobdel:    make(chan string, jobreq_limit),
		pending:   safepending{0, sync.RWMutex{}},
		index:     idx,
		done:      make(chan struct{}),
	}
}

func (w *Worker) PushJob(user string, job *Job) {
	pair := &jobPair{
		key:   user,
		value: job,
	}
	w.jobadd <- pair
}

func (w *Worker) RemoveJob(user string) {
	w.jobdel <- user
}

func (w *Worker) insertJob(key string, value *Job) {
	w.jobqueue[key] = value
	w.pending.Inc()
}

func (w *Worker) deleteJob(key string) {
	delete(w.jobqueue, key)
	w.pending.Dec()
}

func (w *Worker) Run(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fmt.Println("新建一个goruntine，工作池标识：", w.index)
		defer wg.Done()
		ticker := time.NewTicker(time.Second * 60)
		select {
		case data := <-w.broadcast:
			for k, v := range w.jobqueue {
				fmt.Println(data, k, v)
			}
		case jobpair := <-w.jobadd:
			w.insertJob(jobpair.key, jobpair.value)
			fmt.Println("添加任务",jobpair.key)
		case key := <-w.jobdel:
			w.deleteJob(key)
			fmt.Println("删除任务")
		case <-ticker.C:
			fmt.Println("一分钟报警")
		case <-w.done:
			fmt.Println("退出工作池", w.index, "exit")
			break
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.done <- struct{}{}
	}()
}

/*****************线程池实现*****************/
var (
	MaxWorks   = 2
	MaxQueue   = 5 //通道缓存队列数
	MaxSource  = 1
	MaxJobCurd = 1
)

func main() {

	// 初始化工作池
	pool := new(Pool)
	fmt.Println("初始化工作池：", pool)
	wg := &sync.WaitGroup{}
	for i := 0; i < MaxWorks; i++ {
		//创建工作者
		work := NewWorker(i, MaxQueue, MaxSource, MaxJobCurd)
		work.Run(wg)
		//加入工作池
		heap.Push(pool, work)
	}

	fmt.Println("工作池：", pool)

	worker := heap.Pop(pool).(*Worker)
	fmt.Println("获取一个工作组：", worker)
	worker.PushJob("zhangsan", &Job{
		fn: func() int {
			return 120
		},
		result: make(chan string),
	})
	fmt.Println("添加任务,总数：", worker.pending.Get())

	wg.Wait()
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
