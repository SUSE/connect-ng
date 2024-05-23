package connect

import (
	"flag"
	"testing"
)

var testHwinfo = flag.Bool("test-hwinfo", false, "")

func TestGetHwinfo(t *testing.T) {
	if !*testHwinfo {
		t.SkipNow()
	}
	hw, err := getHwinfo()
	t.Logf("HW info: %+v", hw)
	if err != nil {
		t.Fatalf("getHwinfo() failed: %s", err)
	}
	if hw.Hostname == "" {
		t.Error(`Hostname=="", expected not empty`)
	}
	// reading UUID requires root access which is not available in build env
	// if hw.UUID == "" {
	// 	t.Errorf(`UUID=="", expected not empty`)
	// }
	if hw.Cpus == 0 {
		t.Error("Cpus==0, expected>0")
	}
	if hw.Sockets == 0 {
		// on ARM clusters, lscpu can return "Socket(s): -" if DMI is not accessible
		// this parses to hw.Sockets == 0 so we need to skip this test in those cases
		if hw.Arch == archARM && hw.Clusters > 0 {
			t.Log("Reading number of sockets failed on ARM cluster (DMI not accessible?). Check skipped.")
		} else {
			t.Error("Sockets==0, expected>0")
		}
	}
}
