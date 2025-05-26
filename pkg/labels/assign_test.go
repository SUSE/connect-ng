package labels

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/SUSE/connect-ng/internal/testutil"
	"github.com/SUSE/connect-ng/pkg/connection"
)

func TestAssignLabelSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()

	labels := []Label{
		Label{Name: "label1", Description: "label1 description"},
		Label{Name: "label2"},
	}

	// 204 No Content
	response := testutil.Fixture(t, "pkg/labels/assign_labels_success.json")
	body := testutil.Fixture(t, "pkg/labels/assign_labels_body.json")

	conn.On("Do", mock.Anything).Return(response, nil).Run(testutil.MatchBody(t, string(body)))

	fetchedLabels, err := AssignLabels(conn, labels)
	assert.NoError(err)
	assert.Len(fetchedLabels, 2)
	assert.Equal("label2", fetchedLabels[1].Name)

	conn.AssertExpectations(t)
}
