package sst

import (
	"context"
	"fmt"
	"path"
	"reflect"

	arango "github.com/arangodb/go-driver"
	"github.com/pkg/errors"
)

var (
	nilNode = errors.New("sst: node is nil")
)

// Node represents a vertex of a Semantic Spacetime graph
type Node struct {
	// Key is a mandatory field - short name
	Key string `json:"_key"`
	// Data is an arbitrary key value data structure serializable to JSON
	Data map[string]interface{} `json:"data"`
	// Prefix designates node collection origin
	Prefix string
	// Weight is the importance rank
	Weight float64 `json:"weight"`
}

// CreateNode idempotently creates a node of the specified kind
func (s *SST) CreateNode(kind, key string, data map[string]interface{}, weight float64) (*Node, error) {
	return s.createNode(kind+"/", key, data, weight)
}

// MustCreateNode idempotently creates a node of the specified kind, panics on error
func (s *SST) MustCreateNode(kind string, key string, data map[string]interface{}, weight float64) *Node {
	node, err := s.CreateNode(kind, key, data, weight)
	if err != nil {
		panic(err)
	}
	return node
}

// GetNodeData retrieves data of the node for designated key
func (s *SST) GetNodeData(key string) (map[string]interface{}, error) {
	prefix := path.Dir(key)
	rawkey := path.Base(key)

	col, err := s.collectionOf(prefix + "/") // TODO: verify this works
	if err != nil {
		return nil, errors.Wrapf(err, "sst: failed to get node collection for key: %v", key)
	}

	var node Node
	_, err = col.ReadDocument(context.TODO(), rawkey, &node)
	if err != nil {
		return nil, errors.Wrapf(err, "sst: failed to get node for key: %v", key)
	}
	return node.Data, nil
}

// createNode idempotently creates node with the designated prefix
func (s *SST) createNode(prefix string, key string, data map[string]interface{}, weight float64) (*Node, error) {
	node := &Node{
		Data:   data,
		Key:    ToDocumentKey(key),
		Prefix: prefix,
		Weight: weight,
	}
	err := s.insertNode(node)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// insertNode idempotently inserts the node into the collection specified by node.Prefix
func (s *SST) insertNode(node *Node) error {
	nodes, err := s.collectionOf(node.Prefix)
	if err != nil {
		return err
	}
	exists, err := nodes.DocumentExists(context.TODO(), node.Key)
	if err != nil {
		return err
	}
	if !exists {
		_, err := nodes.CreateDocument(context.TODO(), node)
		if err != nil {
			return errors.Wrapf(err, "sst: failed to create node: %v", node)
		}
	} else {
		if node.Data == nil && node.Weight == 0.0 {
			return nil // Do not update the node if there is no data to enter
		}
		var existing Node
		_, err := nodes.ReadDocument(context.TODO(), node.Key, &existing)
		if err != nil {
			return errors.Wrapf(err, "sst: failed to read node: %v", node.Key)
		}
		if existing.Weight != node.Weight || !reflect.DeepEqual(existing.Data, node.Data) {
			_, err := nodes.UpdateDocument(context.TODO(), node.Key, node)
			if err != nil {
				return errors.Wrapf(err, "sst: failed to update node: %v", node)
			}
		}
	}
	return nil
}

// collectionOf identifies node collection based on node prefix
func (s *SST) collectionOf(prefix string) (arango.Collection, error) {
	col := s.nodes[prefix[:len(prefix)-1]]
	if col == nil {
		return nil, errors.New(fmt.Sprintf("sst: no node collection for prefix: %v", prefix))
	}
	return col, nil
}

// NodeID returns the ArangoDB _id for a node
func NodeID(node *Node) (string, error) {
	if node == nil {
		return "", nilNode
	}
	return MustNodeID(node), nil
}

// MustNodeID returns the ArangoDB _id for a node, panics on error
func MustNodeID(node *Node) string {
	return node.Prefix + node.Key
}
