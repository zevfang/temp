package main

import (
	"fmt"
	"container/heap"
)

/*
*************继承container/heap 实现小顶堆的工作池****************
golang标准库中提供了heap结构的容器，可以需要其中的几个方法，来实现一个堆类型的数据结构，
使用时只需要调用标准库中提供的Init初始化接口、Pop接口、Push接口，就可以得到我们想要的结果。
实现的方法有Len、Less、Swap、Push、Pop
*/

type Item struct {
	value    string //任意值&对象
	priority int    // 队列中项的优先级
	index    int    // 堆中项目的索引
}

// Pool实现 heap.Interface，并保存Worker
type PriorityQueue []*Item

func (p PriorityQueue) Len() int {
	return len(p)
}

func (p PriorityQueue) Less(i, j int) bool {
	//最高优先所以使用大于。
	return p[i].priority < p[j].priority
}

func (p PriorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].index = i
	p[j].index = j
}

func (p *PriorityQueue) Push(x interface{}) {
	n := len(*p)
	item := x.(*Item)
	item.index = n
	*p = append(*p, item)
}

func (p *PriorityQueue) Pop() interface{} {
	old := *p
	n := len(*p)
	item := old[n-1]
	*p = old[:n-1]
	return item
}

func main() {
	items := map[string]int{
		"zhangsan": 6, "lisi": 4, "wangwu": 50,
	}

	fmt.Println("初始化队列池")
	pool := make(PriorityQueue, len(items))
	i := 0
	for value, priority := range items {
		pool[i] = &Item{
			value:    value,
			priority: priority,
			index:    i,
		}
		fmt.Println("priority:", pool[i].priority, "--- i:", pool[i].value)
		i++
	}
	heap.Init(&pool)

	fmt.Println("开始按优先级获取值：")
	for pool.Len() > 0 {
		item := heap.Pop(&pool).(*Item)
		fmt.Println("priority:", item.priority, "--- i:", item.value)
	}

}
