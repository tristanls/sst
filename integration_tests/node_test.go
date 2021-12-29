package integration_tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tristanls/sst"
)

func TestCreateNode(t *testing.T) {
	db := arangodb(t)
	defer db.Remove(context.TODO())

	st := st(t)
	n, err := st.CreateNode("Node", "my_node", nil, 1)
	assert.NoError(t, err)

	col, err := db.Collection(context.TODO(), "Node")
	assert.NoError(t, err)
	var stored sst.Node
	_, err = col.ReadDocument(context.TODO(), "my_node", &stored)
	assert.NoError(t, err)

	assert.Equal(t, *n, stored)
}
