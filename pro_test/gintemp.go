package pro_test

import (
	"fmt"
	"runtime"
	"time"
)

var info string
var m = map[int]string{1: "A", 2: "B", 3: "C"}

func main() {
	fmt.Println(&m)
	fmt.Println(info)
}

func init() {
	println("start")
	info = fmt.Sprintf("OS:%s , Arch:%s", runtime.GOOS, runtime.GOARCH, time.Now())
	defer println("end")
}
