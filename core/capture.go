package core

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type CaptureMethod struct {
	Type string

	Envs map[string]string

	//blacklist中出现的文件或文件夹会被忽略
	Blacklist []string

	// 1. "mincores:/" 使用mincores
	Mincores []string

	// 2. "pid:$pid" 分析对应$pid的mapping文件
	PIDs []int

	// 3. "filelist:["$filename"]" 直接传递实际数据．
	FileList []string
	// 3. "filelis:[$includePath"]"
	IncludeList []string

	// 4. "uiapp:wmclass"
	WMClass string

	Processes []string
}

func NewCaptureMethodPIDs(pids ...int) *CaptureMethod {
	return &CaptureMethod{
		Type: _MethodPIDs,
		PIDs: pids,
	}
}

func NewCaptureMethodFileList(fileList []string, includeList []string) *CaptureMethod {
	return &CaptureMethod{
		Type:        _MethodFileList,
		FileList:    fileList,
		IncludeList: includeList,
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
	_MethodProcess  = "process"
)

func DoCapture(m *CaptureMethod, handle FileInfoHandleFunc) error {
	bindMap := getMountBindMap(sysFstabPath)

	switch m.Type {
	case _MethodMincores:
		mps := calcRealTargets(_ReduceFilePath(m.Getenv, m.Mincores...), SystemMountPoints, bindMap)
		return _DoCaptureByMincores(mps, m.wrap(handle), bindMap)
	case _MethodPIDs:
		return _DoCaptureByPIDs(m.PIDs, m.wrap(handle))
	case _MethodProcess:
		return _DoCaptureByProcess(m.Processes, m.wrap(handle))
	case _MethodFileList:
		all := m.FileList
		for _, i := range m.IncludeList {
			all = append(all, ReadFileInclude(i)...)
		}
		return _DoCaptureByFileList(_ReduceFilePath(m.Getenv, all...), true, m.wrap(handle), bindMap)
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

func (m *CaptureMethod) SetEnvs(envs map[string]string) {
	if m.Envs == nil {
		m.Envs = envs
		return
	}
	for k, v := range envs {
		m.Envs[k] = v
	}
}

func (m CaptureMethod) Getenv(key string) string {
	v, ok := m.Envs[key]
	if !ok {
		return os.Getenv(key)
	}
	return v
}

func (m CaptureMethod) wrap(fn FileInfoHandleFunc) FileInfoHandleFunc {
	if len(m.Blacklist) == 0 {
		return fn
	}

	blacklist := _ReduceFilePath(m.Getenv, m.Blacklist...)

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

func _DoCaptureByMincores(mountpoints []string, handle FileInfoHandleFunc, bindMap map[string]string) error {
	if err := supportProduceByKernel(); err != nil {
		return err
	}
	ch := make(chan FileInfo)

	go generateFileInfoByKernel(ch, mountpoints, bindMap)

	for info := range ch {
		if err := handle(info); err != nil {
			return err
		}
	}
	return nil
}

func _DoCaptureByPIDs(pids []int, handle FileInfoHandleFunc) error {
	if len(pids) == 0 {
		return errors.New("not found process")
	}
	fs, err := ReferencedFilesByPID(pids...)
	if err != nil {
		return err
	}
	for _, fname := range fs {
		finfo, err := FileMincore(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ByPIDs FileMincoe %q : %v\n", fname, err)
			continue
		}
		if err := handle(finfo); err != nil {
			return err
		}
	}
	return nil
}

// 调用 pgrep 命令获取和 name 相关的进程号
func getProcessPids(name string, user string) ([]int, error) {
	var args []string
	if user != "" {
		args = append(args, "-U", user)
	}
	// -x 精确匹配命令名称
	args = append(args, "-x", "--", name)
	out, err := exec.Command("pgrep", args...).Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var result []int
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err == nil {
			result = append(result, pid)
		}
	}
	return result, nil
}

func _DoCaptureByProcess(processes []string, handle FileInfoHandleFunc) error {
	var pids []int
	for _, process := range processes {
		tmpPids, _ := getProcessPids(process, "")
		pids = append(pids, tmpPids...)
	}
	return _DoCaptureByPIDs(pids, handle)
}

func _DoCaptureByFileList(list []string, _ bool, handle FileInfoHandleFunc, bindMap map[string]string) error {
	mps := calcRealTargets(list, SystemMountPoints, bindMap)
	fmt.Println("calcRealTargets return mount points:", mps)
	ch := make(chan FileInfo)
	go generateFileInfoByKernel(ch, mps, bindMap)
	for info := range ch {
		for _, i := range list {
			if !strings.HasPrefix(info.Name, i) {
				continue
			}

			if err := handle(info); err != nil {
				return err
			}
		}
	}
	return nil
}
