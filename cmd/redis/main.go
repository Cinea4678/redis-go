package main

import (
	"log"
	"os"
	"redis-go/lib/redis"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Program panicked: %v", r)
			os.Exit(1)
		}
	}()

	redis.Start()
}
