package connect

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockSetLabelsApiCall(t *testing.T, expectedLabels []Label) {
	localSetLabels = func(labels []Label) error {
		assert.ElementsMatch(t, expectedLabels, labels, "setLabels: provided labels do not match expectedLabels")
		return nil
	}
}

func TestAssignAndCreateLabelsOk(t *testing.T) {
	assert := assert.New(t)
	expectedLabels := []Label{
		Label{Name: "label1"},
		Label{Name: "label2"},
	}

	mockSetLabelsApiCall(t, expectedLabels)

	err := AssignAndCrateLabels([]string{"label1", "label2"})
	assert.NoError(err)
}

func TestAssignAndCreateLabelsError(t *testing.T) {
	assert := assert.New(t)

	localSetLabels = func([]Label) error {
		return fmt.Errorf("Cannot set more than 10 labels on system: test-system")
	}

	err := AssignAndCrateLabels([]string{"label1", "label2"})
	assert.ErrorContains(err, "Cannot set more than")
}
