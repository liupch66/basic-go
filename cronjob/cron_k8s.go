package main

import (
	"log"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		log.Println("Hello, world!")
		time.Sleep(time.Second)
	}
}
