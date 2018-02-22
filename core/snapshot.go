package core

import (
	"fmt"
	"sort"
	"sync"
)

type Snapshot struct {
	Infos FileInfos

	cacheLock sync.Mutex
	cache     map[string]bool
}

func (s *Snapshot) Sort() {
	sort.Slice(s.Infos, s.Infos.less)
}

func (s *Snapshot) Add(i FileInfo) error {
	s.cacheLock.Lock()
	if s.cache == nil {
		s.cache = make(map[string]bool)
		for _, info := range s.Infos {
			s.cache[info.Name] = true
		}
	}
	if s.cache[i.Name] {
		s.cacheLock.Unlock()
		return nil
	}
	s.cache[i.Name] = true
	s.cacheLock.Unlock()

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

func (snap *Snapshot) AnalyzeSnapshotLoad(current *Snapshot) (int, float32) {
	cs, _ := snap.sizes()
	p := CompareSnapshot(current, snap).Loaded
	return cs, p
}

func DumpSnapshot(snap *Snapshot) error {
	var totalRAMSize, totalFileSize, totalUsedFileSize int
	var totalFile, usedFile int

	for _, info := range snap.Infos {
		totalFile++
		totalFileSize += int(info.FileSize)
		if len(info.Mapping) > 0 {
			usedFile++
			totalUsedFileSize += int(info.FileSize)
			totalRAMSize += info.RAMSize()
			fmt.Println(info)
		}
	}

	if totalUsedFileSize > 0 {
		fmt.Printf("%s\t%d%%\t%s",
			HumanSize(totalRAMSize),
			totalRAMSize*100/totalUsedFileSize,
			fmt.Sprintf("[FOR %q FILES USED/TOTAL: %d/%d]\n",
				"/",
				usedFile, totalFile,
			),
		)
	}
	return nil
}

type SnapshotDiff struct {
	Added   []FileInfo
	Deleted []FileInfo
	Loaded  float32
}

func (d SnapshotDiff) String() string {
	s := fmt.Sprintf("%d Added, %d Deleted\n", len(d.Added), len(d.Deleted))

	if len(d.Added) > 0 {
		s += fmt.Sprintf("============ %d Added ============\n", len(d.Added))
		for _, v := range d.Added {
			s += fmt.Sprintf("+\t%s\n", v)
		}
		s += fmt.Sprintf("\t%s\n", &Snapshot{Infos: d.Added})

	}
	if len(d.Deleted) > 0 {
		s += fmt.Sprintf("============ %d Deleted ============\n", len(d.Deleted))
		for _, v := range d.Deleted {
			s += fmt.Sprintf("-\t%s\n", v)
		}
		s += fmt.Sprintf("\t%s\n", &Snapshot{Infos: d.Deleted})
	}

	return s
}

func CompareSnapshot(a *Snapshot, b *Snapshot) *SnapshotDiff {
	cache := make(map[string]FileInfo)
	union := &Snapshot{}
	for _, i := range a.Infos {
		cache[i.Name] = i
	}

	var added []FileInfo
	for _, i := range b.Infos {
		if _, ok := cache[i.Name]; !ok {
			added = append(added, i)
		} else {
			union.Add(i)
			delete(cache, i.Name)
		}
	}
	var deleted []FileInfo
	for _, v := range cache {
		deleted = append(deleted, v)
	}

	var loadedPercentage float32
	if totalSize, _ := b.sizes(); totalSize != 0 {
		loadedSize, _ := union.sizes()
		loadedPercentage = float32(loadedSize) / float32(totalSize)
	}

	return &SnapshotDiff{
		added,
		deleted,
		loadedPercentage,
	}
}

func (s *Snapshot) sizes() (int, int) {
	var ret1, ret2 int
	for _, r := range s.Infos {
		ret1 += r.RAMSize()
		ret2 += int(r.FileSize)
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
	snap := &Snapshot{}
	for _, m := range ms {
		if err := DoCapture(m, snap.Add); err != nil {
			return nil, err
		}
	}
	return snap, nil
}
