package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"time"
)

type SnapshotType int

const (
	SnapshotTypeBasic = iota
	SnapshotTypeDesktop
	SnapshotTypeApp
)
const (
	SnapBasic   = "basic"
	SnapDesktop = "desktop"
)

type SnapshotInfo struct {
	IdentifyFile  string
	TakeCounter   int
	TakeTimestamp time.Time
	FileNumbers   int
	RAMsSize      int
	FilesSize     int
}

func (s SnapshotInfo) String() string {
	if s.FileNumbers == 0 {
		return fmt.Sprintf("%q doesn't contain any file", s.IdentifyFile)
	}
	return fmt.Sprintf("%q contains %d files, will occupy %s RAM size, about %d%% of total files, be taken %d times",
		s.IdentifyFile,
		s.FileNumbers,
		humanSize(s.RAMsSize),
		s.RAMsSize*100/(s.FilesSize+1),
		s.TakeCounter,
	)

}

type HistoryDB struct {
	BootTimes int

	BasicInfo   SnapshotInfo
	DesktopInfo SnapshotInfo
	AppsInfo    map[string]SnapshotInfo

	backingFile string
}

func SetupHistoryDB(fname string) *HistoryDB {
	db := &HistoryDB{
		backingFile: fname,
		BasicInfo:   SnapshotInfo{IdentifyFile: SnapBasic},
		DesktopInfo: SnapshotInfo{IdentifyFile: SnapDesktop},
		AppsInfo:    make(map[string]SnapshotInfo),
	}
	f, err := os.Open(fname)
	if err == nil {
		gob.NewDecoder(f).Decode(&db)
	}
	defer f.Close()
	return db
}

func (db *HistoryDB) Save() error {
	f, err := os.Create(db.backingFile)
	if err != nil {
		return err
	}
	return gob.NewEncoder(f).Encode(db)
}

func (db *HistoryDB) UpdateSnapshot(snap *Snapshot, t SnapshotType) {
	switch t {
	case SnapshotTypeBasic:
		db.BasicInfo = UpdateSnapshotInfo(db.BasicInfo, snap)
	case SnapshotTypeDesktop:
		db.DesktopInfo = UpdateSnapshotInfo(db.DesktopInfo, snap)
	case SnapshotTypeApp:
		db.AppsInfo[snap.IdentifyFile] = UpdateSnapshotInfo(db.AppsInfo[snap.IdentifyFile], snap)
	default:
		panic("Unknown Snapshot Type")
	}
}

func UpdateSnapshotInfo(info SnapshotInfo, snap *Snapshot) SnapshotInfo {
	info.TakeCounter++
	info.IdentifyFile = snap.IdentifyFile
	info.FileNumbers = snap.Len()
	info.RAMsSize, info.FilesSize = snap.sizes()
	info.TakeTimestamp = time.Now()
	return info
}

func (db HistoryDB) String() string {
	ret := fmt.Sprintf(
		"The Cache Directory %q has %d snapshots, boot %d times.\n",
		path.Dir(db.backingFile), len(db.AppsInfo), db.BootTimes,
	)
	ret += fmt.Sprintf("----------------Basic Snapshot Info----------\n%v\n%v\n",
		db.BasicInfo, db.DesktopInfo)

	ret += fmt.Sprintf("----------------UI Apps Snapshot Info----------")
	for _, s := range db.AppsInfo {
		ret += fmt.Sprintf("\n%v", s)
	}
	return ret
}
