package main

import (
	"os"
	"fmt"
)

func main() {
	usage := "Usage: todoistist "
	if len(os.Args) < 2 {
		fmt.Println(usage)
	} else if os.Args[0] == "update" {
	} else if os.Args[0] == "config" {
	} else {
	}
}

func setConfig(path string) {
}

func update() {
}

func invoke(name string) {
}
