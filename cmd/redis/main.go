package main

import (
	"flag"
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

	// go run main.go -raof -waof
	redis.ReadAOF = flag.Bool("raof", false, "是否使用aof进行初始化")
	redis.WriteAOF = flag.Bool("waof", false, "是否启动aof协程进行不断持久化")

	redis.Start()
}
