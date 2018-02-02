package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Storeage struct {
}

func StoreTo(fname string, o interface{}) error {
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
