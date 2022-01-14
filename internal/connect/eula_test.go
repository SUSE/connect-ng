package connect

import (
	"strings"
	"testing"
)

var index1 = `directory.yast
license.de.txt
license.es.txt
license.fr.txt
license.it.txt
license.ja.txt
license.ko.txt
license.pt_BR.txt
license.ru.txt
license.txt
license.zh_CN.txt
license.zh_TW.txt
`

func TestParseEULAIndex(t *testing.T) {
	expectedLangs := []string{
		"de", "es", "fr", "it", "ja", "ko",
		"pt_BR", "ru", "en_US", "zh_CN", "zh_TW",
	}
	baseURL := "http://license.server/product1/"
	eulas, err := parseEULAIndex([]byte(index1), baseURL)
	if err != nil {
		t.Fatalf("Parsing error: %+v", err)

	}
	if len(eulas) != len(expectedLangs) {
		t.Fatalf("Unexpected number of items: %v", len(eulas))
	}
	for _, l := range expectedLangs {
		if url, ok := eulas[l]; ok {
			if !strings.HasPrefix(url, baseURL) {
				t.Errorf("Malformed URL found: %v", url)
			}
		} else {
			t.Errorf("Expected language %v not found in index", l)
		}
	}
}

func TestSelectEULALang(t *testing.T) {
	eulas := map[string]string{
		"de":    "de-url",
		"fr":    "fr-url",
		"en":    "en-url",
		"en_US": "en-us-url",
		"pl":    "pl-url",
		"es":    "es-url",
	}
	var tests = []struct {
		input    string
		expected string
	}{
		{"", "en_US"},
		{"C", "en_US"},
		{"POSIX", "en_US"},
		{"de", "de"},
		{"fr.UTF-8", "fr"},
		{"es_BR", "es"},
		{"pl_PL.UTF-8", "pl"},
		{"zh_TW", "en_US"},
	}
	for _, test := range tests {
		CFG.Language = test.input
		out := selectEULALanguage(eulas)
		if out != test.expected {
			t.Errorf("For lang '%v' expected '%v', got '%v'", test.input, test.expected, out)
		}
	}
}

func TestSelectEULALangFallback(t *testing.T) {
	eulas := map[string]string{
		"pl":    "pl-url",
		"zh_TW": "zh-url",
	}
	CFG.Language = "fr_FR.UTF-8"
	out := selectEULALanguage(eulas)
	// first from available ones even if it doesn't make sense for given config
	if out != "pl" {
		t.Errorf("Expected 'pl', got '%v'", out)
	}
}

func TestSelectEULALangEmpty(t *testing.T) {
	eulas := map[string]string{}
	CFG.Language = "pl_PL.UTF-8"
	out := selectEULALanguage(eulas)
	if out != "" {
		t.Errorf("Expected empty, got '%v'", out)
	}
}
