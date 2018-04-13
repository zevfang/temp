package main

import (
	"strconv"
	"fmt"
	"log"
)

func main() {
	i, err := strconv.Atoi("bbbb")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(i)
}
