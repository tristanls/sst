package integration_tests

import (
	"context"
	"testing"

	arango "github.com/arangodb/go-driver"
	"github.com/stretchr/testify/assert"
	"github.com/tristanls/sst"
)

func TestCreateLink(t *testing.T) {
	db := arangodb(t)
	defer db.Remove(context.TODO())
	st := st(t)

	n1, err := st.CreateNode("Node", "from_node", nil, 1)
	assert.NoError(t, err)
	n2, err := st.CreateNode("Node", "to_node", nil, 1)
	assert.NoError(t, err)
	err = st.CreateLink(n1, "near", n2, 1)
	assert.NoError(t, err)

	nearLinks, err := db.Collection(context.TODO(), "Near")
	assert.NoError(t, err)
	var stored sst.Link
	_, err = nearLinks.ReadDocument(context.TODO(), "Node_from_nodenearNode_to_node", &stored)
	assert.NoError(t, err)

	assert.Equal(t, sst.Link{
		Key:    "Node_from_nodenearNode_to_node",
		From:   "Node/from_node",
		To:     "Node/to_node",
		SID:    "near",
		Negate: false,
		Weight: 1,
	}, stored)
}

func TestDeleteLink(t *testing.T) {
	db := arangodb(t)
	defer db.Remove(context.TODO())
	st := st(t)
	nearLinks, err := db.Collection(context.TODO(), "Near")
	containsLinks, err := db.Collection(context.TODO(), "Contains")
	var link sst.Link

	n1, err := st.CreateNode("Node", "from_node", nil, 1)
	assert.NoError(t, err)
	n2, err := st.CreateNode("Node", "to_node", nil, 1)
	assert.NoError(t, err)
	err = st.CreateLink(n1, "near", n2, 1) // n1 --> |near| n2
	assert.NoError(t, err)
	_, err = nearLinks.ReadDocument(context.TODO(), "Node_from_nodenearNode_to_node", &link)
	assert.NoError(t, err)
	err = st.CreateLink(n1, "contains", n2, 1) // n1 --> |near| n2, // n1 --> |contains| n2
	assert.NoError(t, err)
	_, err = containsLinks.ReadDocument(context.TODO(), "Node_from_nodecontainsNode_to_node", &link)
	assert.NoError(t, err)

	err = st.DeleteLink(n1, "near", n2) // n1 --> |contains| n2
	assert.NoError(t, err)
	_, err = nearLinks.ReadDocument(context.TODO(), "Node_from_nodenearNode_to_node", &link)
	assert.True(t, arango.IsNotFound(err))
}
