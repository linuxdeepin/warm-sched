package main

import (
	"fmt"
	"os"
)

func FileExist(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func Log(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
}
