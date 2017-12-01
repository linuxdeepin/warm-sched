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

func TakeSnapshot(dirs []string, fname string) error {
	if !RunByRoot {
		fmt.Fprintln(os.Stderr, "Sorts by disk sector need root privilege. Fallback to sorts by inode")
	}

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

func ShowPlymouthMessage(msg string) {
	cmd := exec.Command("plymouth", "display-message", msg)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("PLY:%v %v\n", err, cmd.Args)
	}
}

func LoadSnapshot(fname string, wait bool, ply bool) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	var snap []SnapshotItem
	err = gob.NewDecoder(f).Decode(&snap)
	if err != nil {
		return err
	}

	n := len(snap)/10 + 1
	for i, r := range snap {
		if ply && i%n == 0 {
			go ShowPlymouthMessage(fmt.Sprintf(`--text=%s -- %d%%`, r.Name, i*10/n))
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
		fmt.Printf("%d %d %s\n", i.inode, i.sector, i.FName)
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
