package core

import (
	"fmt"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

type Finder func(pid int, wmclass string) bool

func NewCaptureMethodUIApp(wmclass string) (*CaptureMethod, error) {
	if wmclass == "" {
		return nil, fmt.Errorf("It must specify wmclass for UIAPP")
	}
	var pid int
	fn := func(_pid int, _wmclass string) bool {
		if _wmclass == wmclass {
			pid = _pid
			return true
		}
		return false
	}
	found := findPidInUIWindows(fn)
	if !found || pid == 0 {
		return nil, fmt.Errorf("Not Found")
	}
	cpath, err := _UIGroupFromPID(pid)
	if err != nil {
		return NewCaptureMethodPIDs(pid), nil
	}
	pids, err := _CGroupPIDs(cpath)
	if err != nil {
		return NewCaptureMethodPIDs(pid), nil
	}
	return NewCaptureMethodPIDs(pids...), nil
}

func _CGroupPIDs(cpath string) ([]int, error) {
	bs, err := ioutil.ReadFile(path.Join(cpath, "cgroup.procs"))
	if err != nil {
		return nil, err
	}
	var ret []int
	for _, str := range strings.Fields(string(bs)) {
		v, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		ret = append(ret, v)
	}
	return ret, nil
}

func _UIGroupFromPID(pid int) (string, error) {
	bs, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(bs), "\n") {
		var noused, sid, uid int
		_, err = fmt.Sscanf(line, "%d:memory:/%d@dde/uiapps/%d", &noused, &sid, &uid)
		if err == nil {
			return fmt.Sprintf("%d@dde/uiapps/%d", sid, uid), nil
		}
	}
	return "", fmt.Errorf("Not Found")
}

func findPidInUIWindows(finder Finder) bool {
	xu, err := xgbutil.NewConnDisplay("")
	if err != nil {
		fmt.Println("findPidInUIWindows:", err)
		return false
	}

	ws, err := xprop.PropValWindows(xprop.GetProperty(xu, xu.RootWin(), "_NET_CLIENT_LIST"))
	if err != nil {
		fmt.Println("findPidInUIWindows:", err)
		return false
	}

	for _, xid := range ws {
		pid, err := xprop.PropValNum(xprop.GetProperty(xu, xid, "_NET_WM_PID"))
		if err != nil || pid == 0 {
			continue
		}
		wm, err := xprop.PropValStrs(xprop.GetProperty(xu, xid, "WM_CLASS"))
		if err != nil || len(wm) != 2 {
			continue
		}
		if finder(int(pid), wm[1]) {
			return true
		}
	}
	return false
}
