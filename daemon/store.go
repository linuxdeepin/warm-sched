package main

import (
	"../core"
	"encoding/gob"
	"fmt"
	"os"
)

func CaptureAndStore(scanMountPoints []string, identifyFile string, fname string) error {
	snap, err := core.CaptureSnapshot(identifyFile, scanMountPoints)
	if err != nil {
		return err
	}
	snap.Sort()
	return storeTo(snap, fname)
}

func LoadAndApply(fname string, ignoreError bool) error {
	snap, err := loadFrom(fname)
	if err != nil {
		return err
	}
	return core.ApplySnapshot(snap, ignoreError)
}

func storeTo(s *core.Snapshot, fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	return gob.NewEncoder(f).Encode(s)
}

func loadFrom(fname string) (*core.Snapshot, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	var snap core.Snapshot
	err = gob.NewDecoder(f).Decode(&snap)
	if err != nil {
		return nil, fmt.Errorf("Invalid snapshot version for %q", fname)
	}
	return &snap, nil
}
