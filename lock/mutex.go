package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var mutex sync.Mutex
	wg := sync.WaitGroup{}

	for i := 1; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			fmt.Println("before lock:", i)
			mutex.Lock()
			time.Sleep(time.Second)
			fmt.Println("locking:", i)
			mutex.Unlock()
			defer wg.Done()
		}(i)
	}
	wg.Wait()

}
