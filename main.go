package main

import "fmt"

type Config struct {
	Organizations []string
	Server        string
	NumWorkers    int
}

func main() {
	fmt.Println("vim-go")
}
