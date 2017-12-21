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

func (m *Manager) TakeSnapshot(identifyFile string) error {
	snap, err := takeSnapshot(identifyFile, m.scanPoints)
	if err != nil {
		return err
	}
	err = m.history.Update(snap)
	if err != nil {
		return err
	}
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

func (m *Manager) LoadSnapshot(identifyFile string) error {
	name := m.identifyFileToPath(identifyFile)
	m.history.Count(identifyFile)
	err := m.history.Save()
	if err != nil {
		return err
	}
	return LoadSnapshot(name, false)
}

func (m *Manager) IncreaseBootTimes() error {
	m.history.BootTimes++
	return m.history.Save()
}

func (m *Manager) ShowHistory() error {
	fmt.Println(m.history)
	return nil
}
