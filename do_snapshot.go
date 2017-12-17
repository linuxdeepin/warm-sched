package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"sort"
)

type Snapshot struct {
	infos []FileCacheInfo
}

type SnapshotItem struct {
	Name   string
	Ranges []MemRange
}

func (i SnapshotItem) String() string {
	ret := fmt.Sprintf("%s\n\t", i.Name)
	for i, r := range i.Ranges {
		if i != 0 {
			ret += ", "
		}
		ret += fmt.Sprintf("[%d,%d]", r.Offset, r.Length)
	}
	return ret
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
		if a.sector == b.sector {
			return a.inode < b.inode
		}
		return a.sector < b.sector
	}
	return a.dev < b.dev
}
func (s *Snapshot) Swap(i, j int) {
	s.infos[i], s.infos[j] = s.infos[j], s.infos[i]
}

func ShowSnapshot(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	var snap []SnapshotItem
	err = gob.NewDecoder(f).Decode(&snap)
	if err != nil {
		return err
	}
	for _, r := range snap {
		fmt.Printf("%v\n", r.Name)
	}
	return nil
}

func TakeSnapshot(dirs []string, fname string) error {
	ch := make(chan FileCacheInfo)
	snap := &Snapshot{}

	Produce(ch, dirs)

	for info := range ch {
		if info.InN == 0 {
			continue
		}
		snap.Add(info)
	}
	sort.Sort(snap)

	return snap.SaveTo(fname)
}

func LoadSnapshot(fname string, wait bool, ply bool) error {
	if _, err := exec.LookPath("plymouth"); err != nil {
		ply = false
	}

	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	var snap []SnapshotItem
	err = gob.NewDecoder(f).Decode(&snap)
	if err != nil {
		return err
	}

	n := len(snap)/20 + 1
	for i, r := range snap {
		if ply && i%n == 0 {
			go ShowPlymouthMessage(fmt.Sprintf("--text=%s -- %d%%", r.Name, i*5/n))
		}

		var err error
		if wait {
			err = Readahead(r.Name, r.Ranges)
		} else {
			err = FAdvise(r.Name, r.Ranges, AdviseLoad)
		}
		if debug {
			fmt.Printf("%+v --> %v\n", r, err)
		}
	}
	return nil
}

func (s *Snapshot) SaveTo(fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}

	var ret []SnapshotItem
	for _, i := range s.infos {
		ret = append(ret, SnapshotItem{i.FName, ToRanges(i.InCache, PageSize64)})
		if debug {
			fmt.Printf("%d %d%% %s\n", i.sector, i.Percentage(), i.FName)
		}
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
