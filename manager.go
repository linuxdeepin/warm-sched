package main

import (
	"fmt"
	"path"
)

type Manager struct {
	scanPoints []string
	cacheDir   string
	history    *HistoryDB
}

func NewManager(mps []string, cacheDir string) (*Manager, error) {
	err := TryMkdir(cacheDir)
	if err != nil {
		return nil, err
	}
	return &Manager{
		cacheDir:   cacheDir,
		scanPoints: mps,
		history:    SetupHistoryDB(path.Join(cacheDir, "history.db")),
	}, nil
}

func (m *Manager) identifyFileToPath(i string) string {
	return path.Join(m.cacheDir, Hash(i)+".snap")
}

func (m *Manager) takeSnapshot(identifyFile string, t SnapshotType) error {
	snap, err := takeSnapshot(identifyFile, m.scanPoints)
	if err != nil {
		return err
	}
	m.history.UpdateSnapshot(snap, t)
	err = m.history.Save()
	if err != nil {
		return err
	}
	return snap.SaveTo(m.identifyFileToPath(snap.IdentifyFile))
}

func (m *Manager) ShowSnapshot(identifyFile string) error {
	name := m.identifyFileToPath(identifyFile)
	return ShowSnapshot(name)
}

func (m *Manager) ShowHistory() error {
	fmt.Println(m.history)
	return nil
}

func (m *Manager) TakeBasic() error {
	return m.takeSnapshot(SnapBasic, SnapshotTypeBasic)
}
func (m *Manager) TakeDesktop() error {
	return m.takeSnapshot(SnapDesktop, SnapshotTypeDesktop)
}
func (m *Manager) TakeApplication(id string) error {
	return m.takeSnapshot(id, SnapshotTypeApp)
}

func (m *Manager) ApplySnapshot(identifyFile string) error {
	name := m.identifyFileToPath(identifyFile)
	return LoadSnapshot(name, false)
}
func (m *Manager) ApplyBasic() error {
	m.history.BootTimes++
	err := m.history.Save()
	if err != nil {
		return err
	}
	return m.ApplySnapshot(SnapBasic)
}
func (m *Manager) ApplyAll() error {
	m.ApplySnapshot(SnapDesktop)

	// Try find which snapshots will be applied
	for id := range m.history.AppsInfo {
		m.ApplySnapshot(id)
	}
	return nil
}
