package core

import (
	"os"
	"path/filepath"
	"testing"
)

type SimpleCaptureResult struct {
	names []string
}

func (r SimpleCaptureResult) Empty() bool {
	return len(r.names) == 0
}
func (r *SimpleCaptureResult) Add(info FileInfo) error {
	r.names = append(r.names, info.Name)
	return nil
}
func (r SimpleCaptureResult) Has(name string) bool {
	for _, i := range r.names {
		if i == name {
			return true
		}
	}
	return false
}

func TestTMPFS(t *testing.T) {
	f := os.ExpandEnv("${XDG_RUNTIME_DIR}/dconf/user")
	_, err := os.Lstat(f)
	if err != nil {
		t.Skip("There hasn't the dconf runtime data")
	}
	info, err := FileMincore(f)
	if err != nil {
		t.Fatal(err)
	}
	if info.Percentage() > 100 {
		t.Log(info.RAMSize(), info.FileSize)
		t.Fatalf("Invalid percentage value %v for %q", info.Percentage(), f)
	}
}

func TestBlackList(t *testing.T) {
	libs, err := filepath.Glob("/lib/*")
	if err != nil {
		t.Fatal(err)
	}

	var list = libs
	list = append(list, "/etc/hosts")
	list = append(list, "/etc/hostname")

	r := SimpleCaptureResult{}
	err = DoCapture(NewCaptureMethodFileList(
		"/lib/x86_64-linux-gnu/libpthread-2.24.so",
		"/etc/hosts",
		"/etc/hostname",
	).Black("/etc", "/"),
		r.Add,
	)
	if err != nil {
		t.Fatal(err)
	}

	if r.Has("/etc/hosts") || r.Has("/etc/hostname") {
		t.Fatal("Result mistmatch")
	}
}

func TestCaptureSelf(t *testing.T) {
	r := SimpleCaptureResult{}
	err := DoCapture(NewCaptureMethodPIDs(os.Getpid()), r.Add)
	if err != nil {
		t.Fatal(err)
	}
	if r.Empty() {
		t.Fatal("At least capture one file")
	}
}
