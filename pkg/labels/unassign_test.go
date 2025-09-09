package labels

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/SUSE/connect-ng/internal/testutil"
	"github.com/SUSE/connect-ng/pkg/connection"
)

func TestUnassignLabelSuccess(t *testing.T) {
	assert := assert.New(t)

	_, conn, _ := connection.NewMockConnectionWithCredentials()

	// 204 No Content
	response := testutil.Fixture(t, "pkg/labels/unassign_label_success.json")
	conn.On("Do", mock.Anything).Return(response, nil).Run(testutil.MatchEmptyBody(t))

	remainingLabels, err := UnassignLabel(conn, 2)
	assert.NoError(err)
	assert.Len(remainingLabels, 1)
	assert.Equal("label1", remainingLabels[0].Name)

	conn.AssertExpectations(t)
}

func TestUnassignLabelUnknownId(t *testing.T) {
	assert := assert.New(t)

	_, conn, _ := connection.NewMockConnectionWithCredentials()

	// 204 No Content
	response := testutil.Fixture(t, "pkg/labels/unassign_label_failed.json")
	conn.On("Do", mock.Anything).Return(response, fmt.Errorf("Couldn't find Label with 'id'=1337"))

	_, err := UnassignLabel(conn, 1337)
	assert.ErrorContains(err, "Couldn't find Label")

	conn.AssertExpectations(t)
}
