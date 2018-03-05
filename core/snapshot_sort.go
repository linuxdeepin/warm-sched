package core

import (
	"sort"
)

func (s *Snapshot) Sort() {
	sort.Sort(s.Infos)

	//sort.Slice(s.Infos, s.Infos.less)
}

func (infos FileInfos) Less(i, j int) bool {
	a, b := infos[i], infos[j]
	if a.dev == b.dev {
		return a.sector < b.sector
	}
	return a.dev < b.dev
}
func (infos FileInfos) Len() int {
	return len(infos)
}
func (infos FileInfos) Swap(i, j int) {
	infos[i], infos[j] = infos[j], infos[i]
}
