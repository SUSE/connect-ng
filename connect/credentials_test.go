package connect

import (
	"testing"
)

func TestLoadFile(t *testing.T) {
	expectUsername := "SCC_ab12ab12ab12ab12ab12ab12ab12ab12"
	expectPassword := "cd34cd34cd34cd34cd34cd34cd34cd34"
	got, _ := LoadFile("../testdata/SCCcredentials")
	if got.Username != expectUsername || got.Password != expectPassword {
		t.Errorf("LoadFile() = %q", got)
	}

}
