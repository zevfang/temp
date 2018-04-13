package main

import (
	"fmt"
	"time"
	"net/http"
	"log"
	"runtime"
	"sync"
)

func main() {
	start := time.Now()

	// ① 正常输出
	// 耗时：491.8469ms
	//for i := 0; i < 10000; i++ {
	//	fmt.Printf("输出：%d \r\n", i)
	//}

	// ② 输出看不出明显差别，几乎没有阻塞,性能也没有提升，系统
	// 耗时：556.893ms
	//runtime.GOMAXPROCS(runtime.NumCPU())
	//var wg sync.WaitGroup
	//for i := 0; i < 10; i++ {
	//	wg.Add(1)
	//	go func(*sync.WaitGroup) {
	//		for j := 0; j < 1000; j++ {
	//			fmt.Printf("输出：%d \r\n", j)
	//		}
	//		defer wg.Done()
	//	}(&wg)
	//}
	//wg.Wait()

	urls := []string{
		"https://github.com/",
		"https://www.baidu.com/",
		"https://www.taobao.com/",
		"https://www.jd.com/",
		"http://www.cnblogs.com/Joetao/articles/5314959.html",
		"http://man.linuxde.net/grep",
		"http://www.cnblogs.com/wuhuacong/p/3317223.html",
	}
	//耗时：11.3682592s
	//耗时：5.489904s
	//耗时：7.2979278s
	//耗时：4.5943712s

	for i := 0; i < len(urls); i++ {
		resp, err := http.Get(urls[i])
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode == 200 {
			log.Printf("访问： %s \r\n", resp.Request.Host)
		} else {
			log.Println("访问失败 \r\n")
		}
	}

	terminal01 := time.Since(start)
	fmt.Printf("耗时：%s \r\n", terminal01)


	start02 := time.Now()
	//耗时：7.1205796s
	//耗时：3.3977401s
	//耗时：1.7842629s
	//耗时：3.5898338s
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	for i := 0; i < len(urls); i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, url string) {
			defer wg.Done()
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			if resp.StatusCode == 200 {
				log.Printf("访问： %s \r\n", resp.Request.Host)
			} else {
				log.Println("访问失败 \r\n")
			}
		}(&wg, urls[i])
	}
	wg.Wait()

	terminal02 := time.Since(start02)
	fmt.Printf("耗时：%s \r\n", terminal02)
}
