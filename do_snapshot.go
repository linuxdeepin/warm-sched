package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"sort"
)

type snapshotItemStatus int

const (
	snapshotItemInvalid snapshotItemStatus = iota
	snapshotItemAlways
	snapshotItemRemoved
)

type Snapshot struct {
	infos  []FileCacheInfo
	status map[string]snapshotItemStatus
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
	snap, err := ParseSnapshot(fname)
	if err != nil {
		return err
	}
	for _, r := range snap {
		fmt.Printf("%v\n", r.Name)
	}
	return nil
}

func ParseSnapshot(fname string) ([]SnapshotItem, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	var snap []SnapshotItem
	err = gob.NewDecoder(f).Decode(&snap)
	if err != nil {
		return nil, err
	}
	return snap, nil
}

func takeSnapshot(mps []string) (*Snapshot, error) {
	ch := make(chan FileCacheInfo)
	snap := &Snapshot{
		status: make(map[string]snapshotItemStatus),
	}
	err := Produce(ch, mps)
	if err != nil {
		return nil, err
	}

	for info := range ch {
		if info.InN == 0 || BlackDirectory.ShouldSkip(info.FName) {
			continue
		}
		snap.Add(info)
	}
	return snap, nil
}

func TakeSnapshot(scanMountPoints []string, fname string) error {
	snap, err := takeSnapshot(scanMountPoints)
	if err != nil {
		return err
	}
	sort.Sort(snap)
	return snap.SaveTo(fname)
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

	n := len(snap)/20 + 1
	for i, r := range snap {
		if BlackDirectory.ShouldSkip(r.Name) {
			continue
		}

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

func (s *Snapshot) ToItems() []SnapshotItem {
	var ret []SnapshotItem
	for _, i := range s.infos {
		if s.status[i.FName] == snapshotItemRemoved {
			continue
		}
		ret = append(ret, SnapshotItem{i.FName, ToRanges(i.InCache, PageSize64)})
	}
	return ret
}

func (s *Snapshot) SaveTo(fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	ret := s.ToItems()
	return gob.NewEncoder(f).Encode(ret)
}

func (s *Snapshot) String() string {
	var str string
	for _, i := range s.infos {
		str += fmt.Sprintf("%s %v\n", i.FName, i.InCache)
	}
	return str
}
