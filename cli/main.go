package main

import (
	"fmt"
	"github.com/jsutton9/preflight/commands"
	"log"
	"os"
)

func main() {
	usage := "Usage: preflight (TEMPLATE_NAME | update | config CONFIG_FILE)"
	logger := log.New(os.Stderr, "", log.Ldate | log.Ltime)
	if len(os.Args) < 2 {
		fmt.Println(usage)
	} else if os.Args[1] == "update" {
		if len(os.Args) == 2 {
			err := commands.Update()
			if err != nil {
				logger.Println("main: error updating: \n\t" + err.Error())
			}
		} else {
			fmt.Println(usage)
		}
	} else if os.Args[1] == "config" {
		if len(os.Args) == 3 {
			err := commands.SetConfig(os.Args[2])
			if err != nil {
				logger.Println("main: error setting config: \n\t" + err.Error())
			}
		} else {
			fmt.Println(usage)
		}
	} else {
		if len(os.Args) == 2 {
			err := commands.Invoke(os.Args[1])
			if err != nil {
				logger.Println("main: error invoking \"" + os.Args[1] + "\": " +
					"\n\t" + err.Error())
			}
		} else {
			fmt.Println(usage)
		}
	}
}
