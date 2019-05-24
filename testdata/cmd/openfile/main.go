package main

import (
	"os"
	"time"
)

func main() {
	path := os.Args[1]
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	time.Sleep(time.Minute)
}
