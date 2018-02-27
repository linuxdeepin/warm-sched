package core

import (
	"testing"
)

func TestX11Clients(t *testing.T) {
	for i := 0; i < 400; i++ {
		err := X11ClientIterate(nil)
		if err != nil {
			t.Fatal(err)
		}
	}
}
