package main

import (
	"fmt"
	"os"
	"strconv"
)

var poolGetter, poolPutter chan string

func startPool() {
	n, _ := strconv.Atoi(os.Getenv("N"))
	if n == 0 {
		n = 1000
	}
	pool := make([]string, 0, n)
	for ; n > 0; n-- {
		pool = append(pool, fmt.Sprintf("/@/%d", n))
	}
	for {
		if len(pool) > 0 {
			select {
			case v := <-poolPutter:
				pool = append(pool, v)
			case poolGetter <- pool[len(pool)-1]:
				pool = pool[:len(pool)-1]
			}
		} else {
			pool = append(pool, <-poolPutter)
		}
	}
}

func init() {
	poolGetter, poolPutter = make(chan string), make(chan string)
	go startPool()
}
