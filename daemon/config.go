package main

import (
	"../core"
	"io/ioutil"
	"path"
)

func LoadConfig(fname string) (*core.SnapshotConfig, error) {
	var cfg core.SnapshotConfig
	err := LoadFrom(fname, &cfg)
	return &cfg, err
}

func ScanConfigs(dir string) ([]*core.SnapshotConfig, error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var ret []*core.SnapshotConfig
	for _, finfo := range fs {
		if finfo.IsDir() {
			continue
		}
		fname := path.Join(dir, finfo.Name())
		v, err := LoadConfig(fname)
		if err != nil {
			return nil, err
		}
		ret = append(ret, v)
	}
	return ret, nil
}
