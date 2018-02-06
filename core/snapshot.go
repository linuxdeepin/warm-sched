package core

import (
	"fmt"
	"sort"
)

type Snapshot struct {
	Infos FileInfos
}

func createSnapshot() *Snapshot {
	snap := &Snapshot{}
	return snap
}

func (s *Snapshot) Sort() {
	sort.Slice(s.Infos, s.Infos.less)
}

func (s *Snapshot) Add(i FileInfo) error {
	s.Infos = append(s.Infos, i)
	return nil
}

func (s *Snapshot) String() string {
	ramSize, fileSize := s.sizes()
	if fileSize == 0 {
		fileSize = 1
	}
	return fmt.Sprintf("%d files occupied %s RAM size, about %d%% of total files",
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
		err = ApplyFileInfo(r)
		if err != nil && !ignoreError {
			return err
		}
	}
	return nil
}

func CaptureSnapshot(ms ...*CaptureMethod) (*Snapshot, error) {
	if len(ms) == 0 {
		return nil, fmt.Errorf("It Must specify at least one Capture methods.")
	}
	snap := createSnapshot()
	for _, m := range ms {
		if err := DoCapture(m, snap.Add); err != nil {
			return nil, err
		}
	}
	return snap, nil
}
