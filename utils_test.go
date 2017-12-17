package main

import (
	"reflect"
	"testing"
)

func TestBitsToMemRange(t *testing.T) {
	ps := PageSize64
	vec := []bool{false, false, true, true, true, false, true, false, true}
	r1 := MemRange{ps * 2, ps * 3}
	r2 := MemRange{ps * 6, ps * 1}
	r3 := MemRange{ps * 8, ps * 1}
	rr := []MemRange{r1, r2, r3}

	MaxAdviseSize = int64(12 * KB)

	r, _ := toRange(vec, ps)
	if !reflect.DeepEqual(r1, r) {
		t.Fatalf("%+v is not equal %+v\n", r1, r)
	}
	if rs := ToRanges(vec, ps); !reflect.DeepEqual(rs, rr) {
		t.Fatalf("\nvec:%v\n%+v\n is not equal\n%+v\n", vec, rs, rr)
	}

	if r := ToRanges(nil, ps); len(r) != 0 {
		t.Fatal("None Empty Set", r)
	}
	if r := ToRanges([]bool{}, ps); len(r) != 0 {
		t.Fatal("None Empty Set", r)
	}

}
