package main

import (
	"log"
	"os"
)

func main() {
	cli := NewCLI(os.Stdout, os.Args[1:])
	if err := cli.Run(); err != nil {
		log.Fatal(err)
	}
}
