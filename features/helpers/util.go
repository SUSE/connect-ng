package helpers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertValidJSON[R any](t *testing.T, jsonString string) R {
	t.Helper()

	var result R

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		assert.FailNow(t, "String is not valid JSON: "+err.Error())
	}
	return result
}
