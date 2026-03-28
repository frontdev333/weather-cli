package main

import (
	"fmt"
	"time"
)

func main() {
	m := make(map[string]int)

	go func() {
		for i := 0; i < 1000; i++ {
			m["counter"] = i
		}
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			m["counter"] = i * 2
		}
	}()

	time.Sleep(time.Second)
	fmt.Println(m["counter"])
}
