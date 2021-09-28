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
no_zypper_refs: true`

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
	expect := Config{BaseURL: "https://smt-azure.susecloud.net", Language: "en_US.UTF-8", NoZypperRefresh: true}
	c := Config{}
	parseConfig(r, &c)
	if !reflect.DeepEqual(c, expect) {
		t.Errorf("got %+v, expected %+v", c, expect)
	}
}

func TestParseConfig2(t *testing.T) {
	r := strings.NewReader(cfg2)
	expect := Config{BaseURL: "http://example.com", Language: "en_US.UTF-8", Insecure: true}
	c := Config{}
	parseConfig(r, &c)
	if !reflect.DeepEqual(c, expect) {
		t.Errorf("got %+v, expected %+v", c, expect)
	}
}

func TestSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SUSEConnect.test")
	c1 := NewConfig()
	c1.Path = path
	if err := c1.Save(); err != nil {
		t.Fatalf("Unable to write config: %s", err)
	}
	c2 := NewConfig()
	c2.Path = path
	c2.Load()
	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("got %+v, expected %+v", c2, c1)
	}
}
