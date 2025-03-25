package connect

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var cfg1 = `---
insecure: false
url: https://smt-azure.susecloud.net
language: en_US.UTF-8
no_zypper_refs: true
auto_agree_with_licenses: true
enable_system_uptime_tracking: false`

var cfg2 = `---
 insecure: true
url :	http://example.com
 language : en_US.UTF-8
# comment
  # indented comment
  # comment with: colon
:
badkey: badval

`

func TestParseConfig(t *testing.T) {
	r := strings.NewReader(cfg1)

	opts := DefaultOptions()
	parseFromConfiguration(r, opts)

	expectedURL := "https://smt-azure.susecloud.net"
	if opts.BaseURL != expectedURL {
		t.Fatalf("Unexpected '%v'; expecting '%v'", opts.BaseURL, expectedURL)
	}
	expectedLanguage := "en_US.UTF-8"
	if opts.Language != expectedLanguage {
		t.Fatalf("Unexpected '%v'; expecting '%v'", opts.Language, expectedLanguage)
	}
	if !opts.NoZypperRefresh {
		t.Fatalf("NoZypperRefresh should be true")
	}
	if !opts.AutoAgreeEULA {
		t.Fatalf("AutoAgreeEULA should be true")
	}
}

func TestParseConfig2(t *testing.T) {
	r := strings.NewReader(cfg2)
	expect := DefaultOptions()
	expect.BaseURL = "http://example.com"
	expect.Language = "en_US.UTF-8"
	expect.Insecure = true

	opts := DefaultOptions()
	parseFromConfiguration(r, opts)
	if !reflect.DeepEqual(opts, expect) {
		t.Errorf("got %+v, expected %+v", opts, expect)
	}
}

func TestSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SUSEConnect.test")
	c1 := DefaultOptions()
	c1.Path = path
	c1.AutoAgreeEULA = true
	c1.ServerType = UnknownProvider
	if err := c1.SaveAsConfiguration(); err != nil {
		t.Fatalf("Unable to write config: %s", err)
	}
	c2, err := ReadFromConfiguration(path)
	if err != nil {
		t.Fatalf("Got an error: %v", err)
	}
	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("got %+v, expected %+v", c2, c1)
	}
}
