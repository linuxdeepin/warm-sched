package core

import (
	"os"
	"testing"
)

func TestX11Clients(t *testing.T) {
	if os.Getenv("DISPLAY") == "" {
		t.Skip("There hasn't X11 environment")
		return
	}
	for i := 0; i < 4000; i++ {
		err := X11ClientIterate(nil)
		if err != nil {
			t.Fatal(err, i)
		}
	}
}
