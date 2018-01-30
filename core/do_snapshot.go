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
	IdentifyFile string
	Inodes       []Inode

	status map[string]snapshotItemStatus
}

func (s *Snapshot) Add(i Inode) {
	s.Inodes = append(s.Inodes, i)
}
func (s *Snapshot) Len() int {
	return len(s.Inodes)
}
func (s *Snapshot) Less(i, j int) bool {
	a, b := s.Inodes[i], s.Inodes[j]

	if a.dev == b.dev {
		return a.sector < b.sector
	}
	return a.dev < b.dev
}
func (s *Snapshot) Swap(i, j int) {
	s.Inodes[i], s.Inodes[j] = s.Inodes[j], s.Inodes[i]
}

func (s *Snapshot) sizes() (int, int) {
	var ret1, ret2 int
	for _, r := range s.Inodes {
		ret1 += r.RAMSize()
		ret2 += int(r.Size)
	}
	return ret1, ret2
}
func (s *Snapshot) String() string {
	ramSize, fileSize := s.sizes()
	if fileSize == 0 {
		fileSize = 1
	}
	return fmt.Sprintf("%q contains %d files, will occupy %s RAM size, about %d%% of total files",
		s.IdentifyFile,
		s.Len(),
		humanSize(ramSize),
		ramSize*100/fileSize,
	)
}

func ShowSnapshot(fname string) error {
	snap, err := ParseSnapshot(fname)
	if err != nil {
		return err
	}
	fmt.Println(snap)
	return nil
}

func ParseSnapshot(fname string) (*Snapshot, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	var snap Snapshot
	err = gob.NewDecoder(f).Decode(&snap)
	if err != nil {
		return nil, fmt.Errorf("Invalid snapshot version for %q", fname)
	}
	return &snap, nil
}

func takeSnapshot(identifyFile string, mps []string) (*Snapshot, error) {
	ch := make(chan Inode)
	snap := &Snapshot{
		IdentifyFile: identifyFile,
		status:       make(map[string]snapshotItemStatus),
	}
	err := Produce(ch, mps)
	if err != nil {
		return nil, err
	}

	for info := range ch {
		if len(info.Mapping) == 0 || BlackDirectory.ShouldSkip(info.Name) {
			continue
		}
		snap.Add(info)
	}
	return snap, nil
}

func TakeSnapshot(scanMountPoints []string, identifyFile string, fname string) error {
	snap, err := takeSnapshot(identifyFile, scanMountPoints)
	if err != nil {
		return err
	}
	return snap.SaveTo(fname)
}

func LoadSnapshot(fname string, wait bool) error {
	snap, err := ParseSnapshot(fname)
	if err != nil {
		return err
	}

	for _, r := range snap.Inodes {
		if BlackDirectory.ShouldSkip(r.Name) {
			continue
		}

		var err error
		if wait {
			err = Readahead(r.Name, r.Mapping)
		} else {
			err = FAdvise(r.Name, r.Mapping, AdviseLoad)
		}
		if debug {
			fmt.Printf("%+v --> %v\n", r, err)
		}
	}
	return nil
}

func (s *Snapshot) ToItems() []Inode {
	if len(s.status) == 0 {
		return s.Inodes
	}
	var ret []Inode
	for _, i := range s.Inodes {
		if s.status[i.Name] == snapshotItemRemoved {
			continue
		}
		ret = append(ret, i)
	}
	return ret
}

func (s *Snapshot) SaveTo(fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	sort.Sort(s)
	return gob.NewEncoder(f).Encode(s)
}
