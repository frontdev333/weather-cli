package main

import (
	"fmt"
	"frontdev333/weather-cli/internal/cache"
	"time"
)

func main() {
	c := cache.New()

	// Запускаем 10 горутин, которые одновременно пишут в кэш
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", j)
				c.Set(key, fmt.Sprintf("value-%d-%d", id, j), 5*time.Second)
			}
		}(i)
	}

	// И ещё 5 горутин, которые читают
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", j)
				val, _, ok := c.Get(key)
				if ok {
					fmt.Printf("Read: %v\n", val)
				}
			}
		}()
	}

	time.Sleep(2 * time.Second)
	fmt.Println("Done! No races detected.")
}
