package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {


	//协程等待
	//future := make(chan string, 1)
	//go func() {
	//	future <- "Hello" //async
	//}()
	//<-future //await


	//获取随机数
	c := make(chan result, 10)
	for i := 0; i < cap(c); i++ {
		go func() {
			v, e := process()
			c <- result{val: v, err: e}
		}()
	}

	for i := 0; i < cap(c); i++ {
		res := <-c
		if res.err == nil {
			fmt.Println(res.val)
		}
	}

}

type result struct {
	val string
	err error
}

func process() (string, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	return string(b[r.Intn(len(b))]), nil
}
