package core

import (
	"fmt"
	"golang.org/x/sys/unix"
	"reflect"
	"syscall"
	"testing"
	"time"
)

func TestRangeStartWithTrues(t *testing.T) {
	vec := []bool{true, true, false, false, false, true}
	ret := []PageRange{
		PageRange{0, 2},
		PageRange{5, 1},
	}

	if r := toRanges(vec); !reflect.DeepEqual(r, ret) {
		t.Fatalf("Expect %v, but got %v", ret, r)
	}
}

func createFile(name string, pageCount int) string {
	return "/usr/bin/git"
}

func randomePageRange(maxCount int) []PageRange {
	return []PageRange{
		PageRange{0, 1},
		PageRange{2, 1},
		PageRange{4, 1},
		PageRange{11, 1},
	}
}

func TestLoad(t *testing.T) {
	a := createFile("A", 50)

	// 1. drop file A
	err := _FAdvise(a, nil, _AdviseDrop)
	if err != nil {
		t.Fatalf("Can't drop %s' page cache", a)
	}
	time.Sleep(time.Second)

	// 2. check result A
	i, err := FileMincore(a)
	if err != nil {
		t.Fatal(err)
	}
	if len(i.Mapping) != 0 {
		t.Fatal("Drop failed", i.Mapping)
	}

	// 3. load random pages
	rs := randomePageRange(50)

	err = _FAdvise(a, rs, _AdviseLoad)
	if err != nil {
		t.Fatal("Failed excute AdviseLoad", err)
	}
	time.Sleep(time.Second)

	i, err = FileMincore(a)
	if err != nil {
		t.Fatal(err)
	}

	// 3. check loaded pages
	if !reflect.DeepEqual(i.Mapping, rs) {
		t.Fatalf("Except %v, but got %v", rs, i.Mapping)
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

	if r := pageRangeToSizeRange(1, 1, d...); !reflect.DeepEqual(r, r1) {
		t.Fatalf("Expect %v, but got %v", r1, r)
	}
	if r := pageRangeToSizeRange(ps, 2, d...); !reflect.DeepEqual(r, r2) {
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
	if rs := toRanges(vec); !reflect.DeepEqual(rs, rr) {
		t.Fatalf("\nvec:%v\n%+v\n is not equal\n%+v\n", vec, rs, rr)
	}

	if r := toRanges(nil); len(r) != 0 {
		t.Fatal("None Empty Set", r)
	}
	if r := toRanges([]bool{}); len(r) != 0 {
		t.Fatal("None Empty Set", r)
	}
	if r := toRanges([]bool{false, false, false}); len(r) != 0 {
		t.Fatal("None Empty Set", r)
	}
}

func TestVerifyMincores(t *testing.T) {
	_DoCaptureByMincores([]string{"/"}, func(info FileInfo) error {
		err := VerifyBySyscall(info)
		if err != nil {
			t.Error(err)
		}
		return nil
	})
}

func VerifyBySyscall(info FileInfo) error {
	if err := syscall.Access(info.Name, unix.R_OK); err != nil {
		return nil
	}
	info2, err := FileMincore(info.Name)
	if err != nil {
		return err
	}
	r1, r2 := info.Mapping, info2.Mapping
	if !reflect.DeepEqual(r1, r2) {
		return fmt.Errorf("WTF: %s \n\tKern:%v(%d)\n\tSys:%v(%d)\n", info2.Name,
			r1, int(info.Size)/pageSize, r2, int(info2.Size)/pageSize)
	}
	return nil
}

func TestEnsureAdviseConst(t *testing.T) {
	if (unix.FADV_DONTNEED != _AdviseDrop) ||
		(unix.FADV_WILLNEED != _AdviseLoad) {
		t.Fatal()
	}
}
