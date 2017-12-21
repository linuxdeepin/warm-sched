package main

import (
	"reflect"
	"testing"
)

func TestRangeStartWithTrues(t *testing.T) {
	vec := []bool{true, true, false, false, false, true}
	ret := []MemRange{
		MemRange{0, 2},
		MemRange{5, 1},
	}

	if r := ToRanges(vec, 32); !reflect.DeepEqual(r, ret) {
		t.Fatalf("Expect %v, but got %v", ret, r)
	}
}

func TestBitsToMemRange(t *testing.T) {
	vec := []bool{false, false, true, true, true, false, true, false, true}
	r1 := MemRange{2, 3}
	r2 := MemRange{6, 1}
	r3 := MemRange{8, 1}
	rr := []MemRange{r1, r2, r3}

	r, _ := toRange(vec, 3)
	if !reflect.DeepEqual(r1, r) {
		t.Fatalf("%+v is not equal %+v\n", r1, r)
	}
	if rs := ToRanges(vec, 3); !reflect.DeepEqual(rs, rr) {
		t.Fatalf("\nvec:%v\n%+v\n is not equal\n%+v\n", vec, rs, rr)
	}

	if r := ToRanges(nil, 0); len(r) != 0 {
		t.Fatal("None Empty Set", r)
	}
	if r := ToRanges([]bool{}, 0); len(r) != 0 {
		t.Fatal("None Empty Set", r)
	}
}
