package main

import (
	"sync"
	"fmt"
)

type User struct {
	UserMap map[string]string
	lock    *sync.RWMutex
}

func NewUserMap() *User {
	return &User{
		UserMap: make(map[string]string),
		lock:    new(sync.RWMutex),
	}
}
func (user *User) Add(key, value string) {
	user.lock.Lock()
	user.UserMap[key] = value
	user.lock.Unlock()
}


func (user *User) Remove(key string) {
	user.lock.Lock()
	delete(user.UserMap, key)
	user.lock.Unlock()
}

func (user *User) Get(key string) string {
	user.lock.RLock()
	m := user.UserMap[key]
	user.lock.RUnlock()
	return m
}

func main() {
	u := NewUserMap()
	fmt.Println(u)
	u.Add("a","bbbbbbb");
	u.Add("zhangsan", "123456")
	s:= u.Get("a")
	fmt.Println(s)
	fmt.Println(u)

	u.Remove("zhangsan")
	fmt.Println(u)
}
