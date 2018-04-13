package main

import (
	"time"
	"runtime"
	"net/http"
	"fmt"
	"log"
	"errors"
)

func main() {

	//urls := []string{
	//	"https://github.com/",
	//	"https://www.baidu.com/",
	//	"https://www.taobao.com/",
	//	"https://www.jd.com/",
	//	"http://www.cnblogs.com/Joetao/articles/5314959.html",
	//	"http://man.linuxde.net/grep",
	//	"http://www.cnblogs.com/wuhuacong/p/3317223.html",
	//}

	//ProNomal(urls)

	//ProChanSync(urls)

	ProChanSort()
}

//耗时：6.4406758s
//普通循环访问
func ProNomal(urls []string) {
	start := time.Now()
	for i := 0; i < len(urls); i++ {
		Get(urls[i])
	}
	terminal := time.Since(start)
	fmt.Printf("耗时：%s \r\n", terminal)
}

//耗时：3.7251586s
//并发同步访问(无缓存)
func ProChanSync(urls []string) {

	start := time.Now()
	runtime.GOMAXPROCS(runtime.NumCPU())

	ch := make(chan bool)

	for i := 0; i < len(urls); i++ {
		go func(url string, c chan bool) {
			if err := Get(url); err == nil {
				fmt.Println(" ok \r\n")
			}
			c <- true
		}(urls[i], ch)
	}

	for i := 0; i < len(urls); i++ {
		select {
		case <-ch:
		case <-time.After(time.Duration(5) * time.Second):
			fmt.Println("time out \r\n")
		}
	}

	terminal := time.Since(start)
	fmt.Printf("耗时：%s \r\n", terminal)
}




//执行顺序返回
func ProChanSort()  {

}



//GET 请求
func Get(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode == 200 {
		log.Printf("访问： %s \r\n", resp.Request.Host)
		return nil
	} else {
		log.Println("访问失败 \r\n")
		return errors.New("错误请求")
	}
}
