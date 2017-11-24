package main

type MemRange struct {
	Offset int64
	Length int64
}

func toRange(vec []bool, pageSize int64) (MemRange, []bool) {
	var s int64
	var offset int64 = -1
	for i, v := range vec {
		if v && offset < 0 {
			offset = int64(i) * pageSize
		}
		if !v && offset > 0 {
			return MemRange{offset, s - offset}, vec[i:]
		}
		s += pageSize
	}
	return MemRange{offset, s - offset}, nil
}

func ToRanges(vec []bool, pageSize int64) []MemRange {
	var ret []MemRange
	var i MemRange
	var pos int64 = -1
	for {
		i, vec = toRange(vec, pageSize)
		if i.Offset == -1 {
			break
		}
		if pos != -1 {
			i.Offset += pos
		}
		pos = i.Offset + i.Length

		ret = append(ret, i)
		if len(vec) == 0 {
			break
		}
	}
	return ret
}
