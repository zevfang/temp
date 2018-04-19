package main

import (
	"fmt"
	"math/rand"
	"time"
)


/*
	实际上该方式就是保证map操作对外只有一个接口，其他线程必须通过该接口才能访问map是，从而实现了线程安全的map
*/
type Item struct {
	NickName string
	Age      int
}

type ItemPair struct {
	key   string
	value *Item
}

type SafeMap struct {
	queue map[string]*Item
	add   chan *ItemPair
	del   chan string
}

func NewSafeMap() *SafeMap {
	m := &SafeMap{
		queue: make(map[string]*Item),
		add:   make(chan *ItemPair),
		del:   make(chan string),
	}
	return m
}

func (s *SafeMap) Insert(k string, v *Item) {
	pair := &ItemPair{
		key:   k,
		value: v,
	}
	s.add <- pair
}

func (s *SafeMap) Remove(key string) {
	s.del <- key
}

func (s *SafeMap) addItem(key string, value *Item) {
	s.queue[key] = value
}

func (s *SafeMap) delItem(key string) {
	delete(s.queue, key)
}

func (s *SafeMap) Run() {
	go func() {
		for {
			select {
			case add := <-s.add:
				s.addItem(add.key, add.value)
			case del := <-s.del:
				s.delItem(del)
			}
		}
	}()
}

func main() {
	m := NewSafeMap()
	m.Run()
	for i := 0; i < 100; i++ {
		go func(n int, dic *SafeMap) {
			item := &Item{
				NickName: fmt.Sprintf("nikename_%d", n),
				Age:      rand.Intn(100),
			}
			m.Insert(fmt.Sprintf("k_%d", n), item)
		}(i, m)
	}
	time.Sleep(time.Second * 3)

	for k, v := range m.queue {
		fmt.Println(k, v)
	}

	//fmt.Println(m)

}
