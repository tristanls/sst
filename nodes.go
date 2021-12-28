package sst

import (
	"context"
	"fmt"
	"path"
	"strings"

	arango "github.com/arangodb/go-driver"
	"github.com/pkg/errors"
)

// Node represents a vertex of a Semantic Spacetime graph
type Node struct {
	// Key is a mandatory field - short name
	Key string `json:"_key"`
	// Data is longer Node description or bulk string data
	Data string `json:"data"`
	// Prefix designates node collection origin
	Prefix string
	// Weight is the importance rank
	Weight float64 `json:"weight"`
}

// CreateNode idempotently creates a node of the specified kind
func (s *SST) CreateNode(short, description, kind string, weight float64) (*Node, error) {
	return s.createNode(short, description, kind+"/", weight)
}

// MustCreateNode idempotently creates a node of the specified kind, panics on error
func (s *SST) MustCreateNode(short, description, kind string, weight float64) *Node {
	node, err := s.CreateNode(short, description, kind, weight)
	if err != nil {
		panic(err)
	}
	return node
}

// GetNodeData retrieves data of the node for designated key
func (s *SST) GetNodeData(key string) (string, error) {
	prefix := path.Dir(key)
	rawkey := path.Base(key)

	col, err := s.collectionOf(prefix + "/") // TODO: verify this works
	if err != nil {
		return "", errors.Wrapf(err, "sst: failed to get node collection for key: %v", key)
	}

	var node Node
	_, err = col.ReadDocument(context.TODO(), rawkey, &node)
	if err != nil {
		return "", errors.Wrapf(err, "sst: failed to get node for key: %v", key)
	}
	return node.Data, nil
}

// createNode idempotently creates node with the designated prefix
func (s *SST) createNode(short, description, prefix string, weight float64) (*Node, error) {
	node := &Node{
		Data:   strings.Trim(description, "\n "),
		Key:    toDocumentKey(short),
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
		if node.Data == "" && node.Weight == 0.0 {
			return nil // Do not update the node if there is no data to enter
		}
		var existing Node
		_, err := nodes.ReadDocument(context.TODO(), node.Key, &existing)
		if err != nil {
			return errors.Wrapf(err, "sst: failed to read node: %v", node.Key)
		}
		if existing != *node {
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
