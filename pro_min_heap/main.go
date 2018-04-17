package main

import (
	"math/rand"
	"fmt"
	"container/heap"
)

/*
*************继承container/heap 实现小顶堆****************
golang标准库中提供了heap结构的容器，可以需要其中的几个方法，来实现一个堆类型的数据结构，
使用时只需要调用标准库中提供的Init初始化接口、Pop接口、Push接口，就可以得到我们想要的结果。
实现的方法有Len、Less、Swap、Push、Pop
*/
type Request struct {
	fn   func() int
	data []byte
	op   int
	c    chan int
}

type Worker struct {
	req     chan Request
	priority int
	index   int
	done    chan struct{}
}

type Pool []*Worker

func (p Pool) Len() int {
	return len(p)
}

func (p Pool) Less(i, j int) bool {
	//最高优先所以使用大于。
	return p[i].priority < p[j].priority
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

var (
	MAXWORKS = 10000
	MAXQUEUE = 1000
)

func main() {
	pool := new(Pool)
	for i := 0; i < 9; i++ {
		work := &Worker{
			req:     make(chan Request, MAXQUEUE),
			priority: rand.Intn(100),
			index:   i,
		}
		fmt.Println("pengding:", work.priority, "--- i:", i)
		heap.Push(pool, work)
	}

	heap.Init(pool)
	fmt.Println("初始化堆成功", pool)

	//for pool.Len() > 0 {
	//	worker := heap.Pop(pool).(*Worker)
	//	fmt.Println("pengding:", worker.pending, "--- i:", worker.index)
	//}

	worker1 := heap.Pop(pool).(*Worker)

	fmt.Println("pengding:", worker1.priority, "--- i:", worker1.index)

}

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

func (w *Worker) Stop()  {
	w.done<- struct{}{}
}

