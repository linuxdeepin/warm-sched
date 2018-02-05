package core

import (
	"fmt"
	"strings"
)

type CaptureMethod struct {
	Type string

	//blacklist中出现的文件或文件夹会被忽略
	Blacklist []string

	// 1. "mincores:/" 使用mincores
	Mincores []string

	// 2. "pid:$pid" 分析对应$pid的mapping文件
	PIDs []int

	// 3. "filelist:["$filename"]" 直接传递实际数据．
	FileList []string

	// 4. "uiapp:wmclass"
	WMClass string
}

func NewCaptureMethodPIDs(pids ...int) *CaptureMethod {
	return &CaptureMethod{
		Type: _MethodPIDs,
		PIDs: pids,
	}
}

func NewCaptureMethodFileList(list ...string) *CaptureMethod {
	return &CaptureMethod{
		Type:     _MethodFileList,
		FileList: list,
	}
}

func NewCaptureMethodMincores(mountPoints ...string) *CaptureMethod {
	return &CaptureMethod{
		Type:     _MethodMincores,
		Mincores: mountPoints,
	}
}

func (m *CaptureMethod) Black(dir ...string) *CaptureMethod {
	//TODO reudce
	m.Blacklist = append(m.Blacklist, dir...)
	return m
}

// Valid type of CaputreMethod
const (
	_MethodMincores = "mincores"
	_MethodPIDs     = "pids"
	_MethodFileList = "filelist"
	_MethodUIApp    = "uiapp"
)

func DoCapture(m *CaptureMethod, handle FileInfoHandleFunc) error {
	switch m.Type {
	case _MethodMincores:
		return _DoCaptureByMincores(m.Mincores, m.wrap(handle))
	case _MethodPIDs:
		return _DoCaptureByPIDs(m.PIDs, m.wrap(handle))
	case _MethodFileList:
		return _DoCaptureByFileList(m.FileList, true, m.wrap(handle))
	case _MethodUIApp:
		_m, err := NewCaptureMethodUIApp(m.WMClass)
		if err != nil {
			return err
		}
		return _DoCaptureByPIDs(_m.PIDs, m.wrap(handle))
	default:
		return fmt.Errorf("Capture method %q is not support", m.Type)
	}
}

func (m CaptureMethod) wrap(fn FileInfoHandleFunc) FileInfoHandleFunc {
	if len(m.Blacklist) == 0 {
		return fn
	}

	blacklist := _ReduceFilePath(m.Blacklist...)

	inBlacklist := func(name string) bool {
		for _, rule := range blacklist {
			if strings.HasPrefix(name, rule) {
				return true
			}
		}
		return false
	}

	return func(finfo FileInfo) error {
		if inBlacklist(finfo.Name) {
			return nil
		}
		return fn(finfo)
	}
}

// real capture methods

func _DoCaptureByMincores(mps []string, handle FileInfoHandleFunc) error {
	if err := supportProduceByKernel(); err != nil {
		return err
	}
	ch := make(chan FileInfo)

	go generateFileInfoByKernel(ch, mps)

	for info := range ch {
		if err := handle(info); err != nil {
			return err
		}
	}
	return nil
}

func _DoCaptureByPIDs(pids []int, handle FileInfoHandleFunc) error {
	fs, err := ReferencedFilesByPID(pids...)
	if err != nil {
		return err
	}
	for _, fname := range fs {
		finfo, err := FileMincore(fname)
		if err != nil {
			continue
		}
		if err := handle(finfo); err != nil {
			return err
		}
	}
	return nil
}

func _DoCaptureByFileList(list []string, _ bool, handle FileInfoHandleFunc) error {
	for _, fname := range list {
		finfo, err := FileMincore(fname)
		if err != nil {
			continue
		}
		if err := handle(finfo); err != nil {
			return err
		}
	}
	return nil
}
