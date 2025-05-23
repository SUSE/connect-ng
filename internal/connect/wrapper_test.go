package connect

import (
	"testing"

	"github.com/SUSE/connect-ng/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAssignLabelsWithSpacesAndNewlines(t *testing.T) {
	assert := assert.New(t)

	conn, _ := testutil.MockConnectionWithCredentials()
	wrapper := Wrapper{
		Connection: conn,
		Registered: true,
	}

	labels := []string{"label-1   ", "\nlabel-2"}
	expected := []labels.Label{
		labels.Label{Name: "label-1"},
		labels.Label{Name: "label-2"},
	}

	body := testutil.Fixture(t, "internal/connect/assign_labels_body.json")
	conn.On("Do").Return(expected, nil).Run(testutil.MatchBody(t, string(body)))

	result, err := wrapper.AssignLabels(labels)
	assert.NoError(err)
	assert.Equal(result, expected)
}
