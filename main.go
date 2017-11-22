package main

import (
	"os"
)

func main() {
	if len(os.Args) == 1 {
		ShowRAMUsage([]string{"."})
	} else {
		ShowRAMUsage(os.Args[1:])
	}
}
