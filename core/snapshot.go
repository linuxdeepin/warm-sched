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
	return fmt.Sprintf("%d files, will occupy %s RAM size, about %d%% of total files",
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

func CaptureSnapshot(cfg CaptureConfig) (*Snapshot, error) {
	if len(cfg.Method) == 0 {
		return nil, fmt.Errorf("It Must specify at least one Capture methods.")
	}
	snap := createSnapshot()
	for _, m := range cfg.Method {
		if err := DoCapture(m, snap.Add); err != nil {
			return nil, err
		}
	}
	return snap, nil
}

type EventSource struct {
	Scope string
	Id    string
}

type SnapshotConfig struct {
	Id          string
	Description string

	// 若IryFile不存在则Apply时会直接忽略
	// 留空或"/"则一直加载，配置为/usr/share/applications/chrome.desktop
	// 之类的则可以避免在chrome已经被卸载的情况下依旧Apply无用的数据.
	TryFile string

	Apply   ApplyConfig
	Capture CaptureConfig
}

type ApplyConfig struct {
	// Usage会通过记录ID对应的EventSource实际发生情况来进行计算
	// InitUsage为Usage的初始值，可以调整静态优先级.
	InitUsage int

	// 列表中所有条目的都被加载后，再进行此次加载
	// 比如UI Apps类型的snapshot都应该等待DE被加载再执行
	After []EventSource

	// 某事件源正在发生时才进行加载
	// 如LaunchRunning, DockRuning, DSCRunning
	In []EventSource
}

type CaptureConfig struct {
	// 小于等于零则, 只会Capture一次
	// 大于零则每次Apply之后对应值减一
	ExpireLimit int

	After []EventSource

	Method []CaptureMethod
}
