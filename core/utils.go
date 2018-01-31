package core

import (
	"fmt"
	"os"
)

var pageSize = os.Getpagesize()
var pageSize64 = int64(pageSize)

func HumanSize(s int) string {
	if s > _GB {
		return fmt.Sprintf("%0.2fG", float32(s)/float32(_GB))
	} else if s > _MB {
		return fmt.Sprintf("%0.1fM", float32(s)/float32(_MB))
	} else if s > _KB {
		return fmt.Sprintf("%0.0fK", float32(s)/float32(_KB))
	} else {
		return fmt.Sprintf("%dB", s)
	}
}

const _KB = 1024
const _MB = 1024 * _KB
const _GB = 1024 * _MB

func roundPageCount(s int64) int {
	return int((s + pageSize64 - 1) / pageSize64)
}

func splitPageRange(r PageRange, mc int) []PageRange {
	var ret []PageRange
	if r.Count <= mc {
		ret = append(ret, r)
	} else {
		i := PageRange{
			Offset: r.Offset,
			Count:  mc,
		}
		j := PageRange{
			Offset: r.Offset + mc,
			Count:  r.Count - mc,
		}
		ret = append(ret, i)
		ret = append(ret, splitPageRange(j, mc)...)
	}
	return ret
}

func pageRangeToSizeRange(pageSize int, maxPageCount int, rs ...PageRange) [][2]int {
	var ret [][2]int
	for _, r := range rs {
		for _, rr := range splitPageRange(r, maxPageCount) {
			ret = append(ret, [2]int{rr.Offset * pageSize, rr.Count * pageSize})
		}
	}
	return ret
}
