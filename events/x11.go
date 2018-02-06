package events

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
)

type X11Source struct {
}

func init() {
	Register(&X11Source{})
}

const x11Scope = "x11"

func (X11Source) Scope() string { return x11Scope }

func (s X11Source) Check(names []string) []string {
	xu, err := xgbutil.NewConnDisplay("")
	if err != nil {
		return nil
	}
	defer xu.Conn().Close()
	ws, err := xprop.PropValWindows(xprop.GetProperty(xu, xu.RootWin(), "_NET_CLIENT_LIST"))
	if err != nil {
		return nil
	}

	var ret []string
	for _, xid := range ws {
		pid, err := xprop.PropValNum(xprop.GetProperty(xu, xid, "_NET_WM_PID"))
		if err != nil || pid == 0 {
			continue
		}
		wm, err := xprop.PropValStrs(xprop.GetProperty(xu, xid, "WM_CLASS"))
		if err != nil || len(wm) != 2 {
			continue
		}
		for _, name := range names {
			if name == wm[1] {
				ret = append(ret, name)
			}
		}
	}
	return ret
}
