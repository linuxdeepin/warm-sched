package main

import (
	"fmt"
	"os"
)

func T(fname string) error {
	info, err := FileMincore(fname)
	if err != nil {
		return err
	}
	if info.Percentage() > 0 {
		fmt.Println(info)
	}
	return nil
}

func main() {
	for _, f := range os.Args[1:] {
		e := T(f)
		if e != nil {
			fmt.Fprintln(os.Stderr, "E:", e)
		}
	}
}
