package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
)

func FileExist(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func EnsureDir(d string) error {
	info, err := os.Stat(d)
	if err == nil && !info.IsDir() {
		return fmt.Errorf("%q is not a directory", d)
	}
	return os.MkdirAll(d, 0755)
}

func Log(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
}

func StoreTo(fname string, o interface{}) error {
	err := EnsureDir(path.Dir(fname))
	if err != nil {
		return err
	}
	w, err := os.Create(fname)
	if err != nil {
		return err
	}
	return storeTo(w, o)
}

func LoadFrom(fname string, o interface{}) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	err = loadFrom(f, o)
	if err != nil {
		return fmt.Errorf("LoadFrom(%q, %T) -> %q", fname, o, err.Error())
	}
	return nil
}

func storeTo(w io.Writer, o interface{}) error {
	return json.NewEncoder(w).Encode(o)
}
func loadFrom(r io.Reader, o interface{}) error {
	return json.NewDecoder(r).Decode(o)
}
