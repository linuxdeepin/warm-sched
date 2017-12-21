package main

import (
	"reflect"
	"testing"
)

func TestRangeStartWithTrues(t *testing.T) {
	vec := []bool{true, true, false, false, false, true}
	ret := []PageRange{
		PageRange{0, 2},
		PageRange{5, 1},
	}

	if r := ToRanges(vec); !reflect.DeepEqual(r, ret) {
		t.Fatalf("Expect %v, but got %v", ret, r)
	}
}

func TestMaxRange(t *testing.T) {
	d := []PageRange{
		PageRange{0, 2},
		PageRange{5, 1},
	}
	r1 := [][2]int{
		[2]int{0, 1},
		[2]int{1, 1},
		[2]int{5, 1},
	}
	ps := 100
	r2 := [][2]int{
		[2]int{0 * ps, 2 * ps},
		[2]int{5 * ps, 1 * ps},
	}

	if r := PageRangeToSizeRange(1, 1, d...); !reflect.DeepEqual(r, r1) {
		t.Fatalf("Expect %v, but got %v", r1, r)
	}
	if r := PageRangeToSizeRange(ps, 2, d...); !reflect.DeepEqual(r, r2) {
		t.Fatalf("Expect %v, but got %v", r2, r)
	}
}

func TestBitsToPageRange(t *testing.T) {
	vec := []bool{false, false, true, true, true, false, true, false, true}
	r1 := PageRange{2, 3}
	r2 := PageRange{6, 1}
	r3 := PageRange{8, 1}
	rr := []PageRange{r1, r2, r3}

	r, _ := toRange(vec)
	if !reflect.DeepEqual(r1, r) {
		t.Fatalf("%+v is not equal %+v\n", r1, r)
	}
	if rs := ToRanges(vec); !reflect.DeepEqual(rs, rr) {
		t.Fatalf("\nvec:%v\n%+v\n is not equal\n%+v\n", vec, rs, rr)
	}

	if r := ToRanges(nil); len(r) != 0 {
		t.Fatal("None Empty Set", r)
	}
	if r := ToRanges([]bool{}); len(r) != 0 {
		t.Fatal("None Empty Set", r)
	}
}
