package main

import "fmt"

type Person struct {
	Name string
	Age  int
}

func main() {

	ProMake()
	//ProNew()
}

/*
	make(T, args) 返回的是 T 的 引用(结构实例) 非指针
	返回传入类型的引用(结构实例) 非指针

	注：make 只能用于 slice,map,channel，引用类型也只针对以上三种类型

*/
func ProMake() {

	//slice 初始化不要给长度，不然则变成数组（slice底层是数组）
	var s []string
	fmt.Println(s)

	s1 := make([]int, 10)
	fmt.Println(s1)

	//map 初始化不需要长度，因为其长度可变
	var m = map[string]Person{}
	fmt.Println(m)

	m1 := make(map[string]Person)
	fmt.Println(m1)

	//chan 初始化 缓存和非缓存channel
	var c = make(chan bool, 10)
	fmt.Println(c)

	c1 := make(chan bool)
	fmt.Println(c1)

	c2 := make(chan bool, 10)
	fmt.Println(c2)

}

/*
	new(T) 返回的是 T 的指针
	返回新分配类型0值的指针

	特点：一般比较少用到
	备注：涉及到寻址的问题，理论上来讲 x 的字段 m，调用上 x.m 等价于 (&x).m
*/
func ProNew() {

	// init
	p := new(string)
	fmt.Println(&p)

	var p1 Person
	fmt.Printf("p1 is %T", p1)
	fmt.Println(p1)

	var p2 = Person{}
	fmt.Printf("p2 is %T", p2)
	fmt.Println(p2)

	var p3 = &Person{}
	fmt.Printf("p3 is %T", p3)
	fmt.Println(p3)

	var p4 = new(Person)
	fmt.Printf("p4 is %T", p4)
	fmt.Println(p4)

	var p5 *Person = new(Person)
	fmt.Printf("p5 is %T", p5)
	fmt.Println(p5)

}

/*
	小结
	new(T) 返回 T 的指针 *T 并指向 T 的零值。
	make(T) 返回的初始化的 T，只能用于 slice，map，channel。
*/
