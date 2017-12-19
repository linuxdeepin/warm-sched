package main

import (
	"fmt"
	"path"
	"path/filepath"
)

func runAndRecordOpenFiles(idFile string, cgroup string) ([]string, error) {
	if !FileExists(idFile) {
		return nil, fmt.Errorf("Can't find identify file %q", idFile)
	}
	return nil, nil
}

func EnumerateAllApps(cacheDir string) []string {
	all, _ := filepath.Glob(path.Join(cacheDir, "apps", "*"))
	return all
}

func TakeApplicationSnapshot(cacheDir string, scans []string, identiFile string) error {
	files, err := runAndRecordOpenFiles(identiFile, "")
	if err != nil {
		return err
	}

	app, err := takeSnapshot(scans)
	if err != nil {
		return err
	}
	fmt.Printf("Collected %d files for %q\n", len(app.infos), identiFile)
	app.Always(files)
	fmt.Printf("Mark %d files as directly dependence\n", len(files))

	reduceSnapshot(app, path.Join(cacheDir, SnapFull))

	fname := path.Join(cacheDir, "apps", path.Base(identiFile))
	for _, other := range EnumerateAllApps(cacheDir) {
		if other == fname {
			continue
		}
		reduceSnapshot(app, other)
	}

	items := app.ToItems()
	if len(items) == 0 {
		fmt.Printf("All files exists in other snapshots")
		return nil
	} else {
		fmt.Printf("Actually take %d files to %s\n", len(app.ToItems()), fname)
		return app.SaveTo(fname)
	}
}

func reduceSnapshot(snap *Snapshot, base string) error {
	baseItems, err := ParseSnapshot(base)
	if err != nil || len(baseItems) == 0 {
		return fmt.Errorf("Can't load snapshot of %s. Application snapshot should be take on top of it. E:%s", base, err)
	}

	var removed int
	for _, f := range baseItems {
		if snap.Remove(f.Name) {
			removed++
		}
	}
	fmt.Printf("Removed %d files because they exists in %s\n", removed, base)
	return nil
}

func (s *Snapshot) Always(names []string) {
	for _, name := range names {
		s.status[name] = snapshotItemAlways
	}
}

func (s *Snapshot) Remove(name string) bool {
	if s.status[name] == snapshotItemAlways {
		return false
	}
	s.status[name] = snapshotItemRemoved
	return true
}
