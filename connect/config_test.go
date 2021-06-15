package connect

import (
	"strings"
	"testing"
)

var cfg1 = `---
insecure: false
url: https://smt-azure.susecloud.net
language: en_US.UTF-8`

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
	expect := Config{"", "https://smt-azure.susecloud.net", "en_US.UTF-8", false}
	c := Config{}
	parseConfig(r, &c)
	if !c.equals(expect) {
		t.Errorf("got %+v, expected %+v", c, expect)
	}
}

func TestParseConfig2(t *testing.T) {
	r := strings.NewReader(cfg2)
	expect := Config{"", "http://example.com", "en_US.UTF-8", true}
	c := Config{}
	parseConfig(r, &c)
	if !c.equals(expect) {
		t.Errorf("got %+v, expected %+v", c, expect)
	}
}
