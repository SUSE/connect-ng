package connect

import (
	"fmt"
	"os"
	"testing"
)

func TestParseJSON(t *testing.T) {
	json, _ := os.ReadFile("../testdata/activations.json")
	got := ParseJSON(json)
	// TODO finish
	fmt.Printf("%+v\n", got)
}
