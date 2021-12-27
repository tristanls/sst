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
	// Prefix designates collection origin, e.g.: Hub, Node, Fragment
	Prefix string
	// Weight is the importance rank
	Weight float64 `json:"weight"`
}

// CreateFragment idempotently creates a Fragment node
func (s *SST) CreateFragment(short, description string, weight float64) (*Node, error) {
	return s.createNode(short, description, "Fragments/", weight)
}

// CreateNode idempotently creates a Node node
func (s *SST) CreateNode(short, description string, weight float64) (*Node, error) {
	return s.createNode(short, description, "Nodes/", weight)
}

// CreateHub idempotently creates a Hub node
func (s *SST) CreateHub(short, description string, weight float64) (*Node, error) {
	return s.createNode(short, description, "Hubs/", weight)
}

// GetNodeData retrieves data of the node for designated key
func (s *SST) GetNodeData(key string) (string, error) {
	scale := path.Dir(key)
	rawkey := path.Base(key)

	col, err := s.collectionOf(scale)
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
		Key:    short,
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

// collectionOf identifies node collection based on scale needed
func (s *SST) collectionOf(scale string) (arango.Collection, error) {
	switch scale {
	case "Hubs/":
		return s.hubs, nil
	case "Nodes/":
		return s.nodes, nil
	case "Fragments/":
		return s.frags, nil
	}
	return nil, errors.New(fmt.Sprintf("sst: no node collection for scale: %v", scale))
}
