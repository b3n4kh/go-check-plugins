package checksystemd

import (
	"testing"
)

func TestSystemd(t *testing.T) {
	sysstate, err := getSystemd()
	if err != nil {
		t.Fatal(err)
	}
	if len(sysstate) > 0 {
		t.Fatal("No error")
	}
}
