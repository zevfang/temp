package main

import (
	"fmt"
)
/*********defer 注意事项***********/
/*
	① 官方声明defer是在return之前执行的
	② 值类型和引用类型问题
	③ 值类型：
			  1.defer 函数未传值，且内部对外部变量赋值，返回值原始值
              2.defer 函数传值，且内部对外部变量赋值，返回值依然是原始值
	   引用型：
   			  1.defer 函数未传值，且内部对外部变量赋值，返回值原始数据
              2.defer 函数传值，且内部对外部引用变量赋值，返回值如果是值类型则为原始数据，复杂类型返回defer赋值结果
	④ 执行顺序问题
		return 非原子性语法糖，分两步执行 1.赋值  2.返回

*/
func main() {
	i:=0
	fmt.Println(&i)
	
	fmt.Println(aaa())
}

type User struct {
	Name string
}

func aaa() string {

	user := &User{}
	user.Name = "lisi"
	fmt.Println(user.Name)

	defer func(user *User) {
		user.Name = "zhangsan"
		fmt.Println(user.Name)
	}(user)
	return user.Name
	//return user
}
