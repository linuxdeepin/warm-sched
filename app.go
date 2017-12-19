package main

import (
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
)

func runAndRecordOpenFiles(prog string) ([]string, error) {
	err := exec.Command(prog).Run()
	return nil, err
}

func enumerateAllApps(cacheDir string) []string {
	all, _ := filepath.Glob(path.Join(cacheDir, "apps", "*"))
	return all
}

func TakeApplicationSnapshot(cacheDir string, prog string) error {
	files, err := runAndRecordOpenFiles(prog)
	if err != nil {
		return err
	}

	app, err := takeSnapshot(ListMountPoints())
	if err != nil {
		return err
	}
	fmt.Printf("Collected %d files after executed %q\n", len(app.infos), prog)
	app.Always(files)
	fmt.Printf("Mark %d files as directly dependence\n", len(files))

	reduceSnapshot(app, path.Join(cacheDir, SnapFull))

	fname := path.Join(cacheDir, "apps", path.Base(prog))
	for _, other := range enumerateAllApps(cacheDir) {
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
