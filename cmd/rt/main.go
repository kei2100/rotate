package main

import (
	"bufio"
	"os"

	"github.com/kei2100/rotate"
)

func main() {
	w, err := rotate.NewWriter("/tmp/rotate", "log.log", rotate.WithConfigFunc(rotate.SizeBasedConfig(2)))
	if err != nil {
		panic(err)
	}
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		w.Write([]byte(stdin.Text()))
	}
}
