package main

import (
	"../core"
	"io/ioutil"
	"path"
)

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
	After []string

	// 某事件源正在发生时才进行加载
	// 如LaunchRunning, DockRuning, DSCRunning
	Before []string
}

type CaptureConfig struct {
	After  []string
	Before []string

	AlwaysLoad bool

	WaitSeond int

	Method []*core.CaptureMethod
}

func LoadConfig(fname string) (*SnapshotConfig, error) {
	var cfg SnapshotConfig
	err := core.LoadJSONFrom(fname, &cfg)
	return &cfg, err
}

func ScanConfigs(dir string) ([]*SnapshotConfig, error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var ret []*SnapshotConfig
	for _, finfo := range fs {
		if finfo.IsDir() {
			continue
		}
		baseName := finfo.Name()
		if path.Ext(baseName) != ".json" {
			continue
		}
		fname := path.Join(dir, baseName)
		v, err := LoadConfig(fname)
		if err != nil {
			return nil, err
		}
		if v.TryFile != "" && !FileExist(v.TryFile) {
			Log("Ignore %q because %q doesn't exists\n", fname, v.TryFile)
			continue
		}
		ret = append(ret, v)

	}
	return ret, nil
}

func (d *Daemon) LoadConfigs(dir string) error {
	cfgs, err := ScanConfigs(dir)
	if err != nil {
		return err
	}
	d.cfgs = cfgs
	return nil
}
