package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"sort"
)

type Snapshot struct {
	infos []FileCacheInfo
}

func (s *Snapshot) Add(i FileCacheInfo) {
	s.infos = append(s.infos, i)
}
func (s *Snapshot) Len() int {
	return len(s.infos)
}
func (s *Snapshot) Less(i, j int) bool {
	a, b := s.infos[i], s.infos[j]

	if a.dev == b.dev {
		return a.sector < b.sector
	}
	return a.dev < b.dev
}
func (s *Snapshot) Swap(i, j int) {
	s.infos[i], s.infos[j] = s.infos[j], s.infos[i]
}

func TakeSnapshot(dirs []string, fname string) error {
	ch := make(chan FileCacheInfo)
	snap := &Snapshot{}

	go Produce(ch, dirs)

	for info := range ch {
		if info.InN == 0 {
			continue
		}
		snap.Add(info)
	}
	sort.Sort(snap)

	return snap.SaveTo(fname)
}

func LoadSnapshot(fname string, wait bool) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	var snap []SnapshotItem
	err = gob.NewDecoder(f).Decode(&snap)
	if err != nil {
		return err
	}
	for _, i := range snap {
		var err error
		if wait {
			err = Readahead(i.Name, i.Ranges)
		} else {
			err = FAdvise(i.Name, i.Ranges, AdviseLoad)
		}
		if debug {
			fmt.Printf("%+v --> %v\n", i, err)
		}
	}
	return nil
}

type SnapshotItem struct {
	Name   string
	Ranges []MemRange
}

func (s *Snapshot) SaveTo(fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}

	var ret []SnapshotItem
	for _, i := range s.infos {
		ret = append(ret, SnapshotItem{i.FName, ToRanges(i.InCache, PageSize64)})
		fmt.Printf("%d %s\n", i.sector, i.FName)
	}

	return gob.NewEncoder(f).Encode(ret)
}

func (s *Snapshot) String() string {
	var str string
	for _, i := range s.infos {
		str += fmt.Sprintf("%s %v\n", i.FName, i.InCache)
	}
	return str
}
