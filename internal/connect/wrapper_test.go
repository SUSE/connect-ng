package connect

import (
	"testing"

	"github.com/SUSE/connect-ng/internal/testutil"
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/labels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssignLabelsWithSpacesAndNewlines(t *testing.T) {
	assert := assert.New(t)

	_, conn, _ := connection.NewMockConnectionWithCredentials()
	wrapper := Wrapper{
		Connection: conn,
		Registered: true,
	}

	toAssign := []string{"label-1   ", "\nlabel-2"}
	expected := []labels.Label{
		labels.Label{Id: 1, Name: "label-1"},
		labels.Label{Id: 2, Name: "label-2"},
	}

	body := testutil.Fixture(t, "internal/connect/assign_labels_body.json")
	response := testutil.Fixture(t, "internal/connect/assign_labels_response.json")
	conn.On("Do", mock.Anything).Return(response, nil).Run(testutil.MatchBody(t, string(body)))

	result, err := wrapper.AssignLabels(toAssign)
	assert.NoError(err)
	assert.Equal(result, expected)
}
