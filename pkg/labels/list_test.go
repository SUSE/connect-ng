package labels

import (
	"testing"

	"github.com/SUSE/connect-ng/internal/testutil"
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListLabels(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()

	// 204 No Content
	response := testutil.Fixture(t, "pkg/labels/list_labels_success.json")

	conn.On("Do", mock.Anything).Return(response, nil).Run(testutil.MatchEmptyBody(t))

	labels, err := ListLabels(conn)
	assert.NoError(err)
	assert.Len(labels, 3)
	assert.Equal("label2", labels[1].Name)

	conn.AssertExpectations(t)
}
