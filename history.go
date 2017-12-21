package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"time"
)

type SnapshotHistory struct {
	IdentifyFile      string
	UsageCounter      int
	LastUsedTimestamp time.Time
	TakeTimestamp     time.Time
	FileNumbers       int
	RAMsSize          int
	FilesSize         int
}

func (s SnapshotHistory) String() string {
	return fmt.Sprintf("%q contains %d files, will occupy %s RAM size, about %d%% of total files, Usage %d times",
		s.IdentifyFile,
		s.FileNumbers,
		humanSize(s.RAMsSize),
		s.RAMsSize*100/(s.FilesSize+1),
		s.UsageCounter,
	)

}

type HistoryDB struct {
	BootTimes   int
	Infos       map[string]SnapshotHistory
	backingFile string
}

func SetupHistoryDB(fname string) *HistoryDB {
	db := &HistoryDB{
		backingFile: fname,
		Infos:       make(map[string]SnapshotHistory),
	}
	f, err := os.Open(fname)
	if err == nil {
		gob.NewDecoder(f).Decode(&db)
	}
	defer f.Close()
	return db
}

func (db HistoryDB) Save() error {
	f, err := os.Create(db.backingFile)
	if err != nil {
		return err
	}
	return gob.NewEncoder(f).Encode(db)
}

func (db HistoryDB) Count(id string) error {
	info, ok := db.Infos[id]
	if !ok {
		return fmt.Errorf("Snapshot %q doesn't exists", id)
	}
	info.UsageCounter++
	info.LastUsedTimestamp = time.Now()
	db.Infos[id] = info
	return nil
}

func (db HistoryDB) Update(snap *Snapshot) error {
	info := db.Infos[snap.IdentifyFile]
	info.IdentifyFile = snap.IdentifyFile
	info.FileNumbers = snap.Len()
	info.RAMsSize, info.FilesSize = snap.sizes()
	info.TakeTimestamp = time.Now()
	db.Infos[snap.IdentifyFile] = info
	return nil
}

func (db HistoryDB) String() string {
	ret := fmt.Sprintf(
		"The Cache Directory %q has %d snapshots, found %d boot times.",
		path.Dir(db.backingFile), len(db.Infos), db.BootTimes,
	)
	for _, s := range db.Infos {
		ret += fmt.Sprintf("\n%v", s)
	}
	return ret
}
