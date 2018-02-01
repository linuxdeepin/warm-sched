package core

import (
	"fmt"
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
	Infos        FileInfos

	status map[string]snapshotItemStatus
}

func createSnapshot(idFile string) *Snapshot {
	snap := &Snapshot{
		IdentifyFile: idFile,
		status:       make(map[string]snapshotItemStatus),
	}
	return snap
}

func (s *Snapshot) Sort() {
	sort.Slice(s.Infos, s.Infos.less)
}

func (s *Snapshot) Add(i FileInfo) {
	s.Infos = append(s.Infos, i)
}

func (s *Snapshot) String() string {
	ramSize, fileSize := s.sizes()
	if fileSize == 0 {
		fileSize = 1
	}
	return fmt.Sprintf("%q contains %d files, will occupy %s RAM size, about %d%% of total files",
		s.IdentifyFile,
		len(s.Infos),
		HumanSize(ramSize),
		ramSize*100/fileSize,
	)
}

func (s *Snapshot) sizes() (int, int) {
	var ret1, ret2 int
	for _, r := range s.Infos {
		ret1 += r.RAMSize()
		ret2 += int(r.Size)
	}
	return ret1, ret2
}

type FileInfos []FileInfo

func (infos FileInfos) less(i, j int) bool {
	a, b := infos[i], infos[j]
	if a.dev == b.dev {
		return a.sector < b.sector
	}
	return a.dev < b.dev
}

func ApplySnapshot(snap *Snapshot, ignoreError bool) error {
	for _, r := range snap.Infos {
		var err error
		err = Apply(r)
		if err != nil && !ignoreError {
			return err
		}
	}
	return nil
}

func CaptureSnapshot(identifyFile string, mps []string) (*Snapshot, error) {
	snap := createSnapshot(identifyFile)
	err := CaptureByMincores(mps, func(info FileInfo) error {
		snap.Add(info)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return snap, nil
}
