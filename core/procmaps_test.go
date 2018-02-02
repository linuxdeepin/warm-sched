package core

import (
	"os"
	"testing"
)

func TestReferencedFilesByPID(t *testing.T) {
	fs, err := ReferencedFilesByPID(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Referenced %d files %v\n", len(fs), fs)
	if len(fs) == 0 {
		t.Fatal("Can't find any referenced file")
	}
}
