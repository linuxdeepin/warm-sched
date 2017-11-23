package main

import (
	"reflect"
	"testing"
)

func TestBitsToMemRange(t *testing.T) {
	ps := PageSize64
	vec := []bool{false, false, true, true, true, false}
	result := MemRange{ps * 2, ps * 3}

	r, _ := toRange(vec, ps)
	if !reflect.DeepEqual(result, r) {
		t.Fatalf("%+v is not equal %+v\n", result, r)
	}
}
