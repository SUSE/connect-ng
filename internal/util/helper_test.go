package util

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestThrowErrorIfEmptyWithNonEmptyString(t *testing.T) {
	result, err := ThrowErrorIfEmpty("testToken")
	assert.NoError(t, err)
	assert.Equal(t, "testToken", result)
}

func TestThrowErrorIfEmptyWithEmptyString(t *testing.T) {
	result, err := ThrowErrorIfEmpty("")
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, fmt.Errorf("string is empty"), err)
}

func TestThrowErrorIfEmptyWithWhitespace(t *testing.T) {
	result, err := ThrowErrorIfEmpty(" ")
	assert.NoError(t, err)
	assert.Equal(t, " ", result)
}

func TestThrowErrorIfEmptyWithNonWhitespace(t *testing.T) {
	result, err := ThrowErrorIfEmpty("\t\n\r")
	assert.NoError(t, err)
	assert.Equal(t, "\t\n\r", result)
}
