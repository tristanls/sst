// Package sst provides facilities for modeling Semantic Spacetime
// in an ArangoDB.
package sst

import (
	"context"
	"regexp"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/pkg/errors"
)

var (
	startEvent = &Node{Key: "start"}
)

// Config contains initial configuration for a Semantic Spacetime
type Config struct {
	// Associations, if specified, will override the default associations for this SST
	Associations map[string]*Association
	Name         string
	// NodeCollections are the names of node collections to instantiate for this SST
	NodeCollections []string
	Password        string
	URL             string
	Username        string
}

type SST struct {
	associations map[string]*Association
	db           arango.Database
	client       arango.Client
	config       *Config
	conn         arango.Connection
	graph        arango.Graph
	name         string

	nodes map[string]arango.Collection

	follows   arango.Collection
	contains  arango.Collection
	expresses arango.Collection
	near      arango.Collection

	prevEvents []*Node
}

var (
	keyRegex = regexp.MustCompile(`[^a-zA-Z0-9_:.@()+,=;$!*'%-]`)
)

// Creates new Semantic Spacetime model backed by ArangoDB
func NewSST(config *Config) (*SST, error) {
	sst := &SST{
		config: config,
		name:   "semantic_spacetime",
	}

	if config.Associations != nil {
		sst.associations = config.Associations
	} else {
		sst.associations = make(map[string]*Association)
		for k, v := range associations {
			sst.associations[k] = v
		}
	}

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{config.URL},
	})
	if err != nil {
		return nil, errors.Wrap(err, "sst: failed to create ArangoDB connection")
	}
	sst.conn = conn
	client, err := arango.NewClient(arango.ClientConfig{
		Connection:     sst.conn,
		Authentication: arango.BasicAuthentication(config.Username, config.Password),
	})
	if err != nil {
		return nil, errors.Wrap(err, "sst: failed to create ArangoDB client")
	}
	sst.client = client

	sst.db, err = sst.openDatabase(sst.config.Name)
	if err == databaseDoesNotExist {
		sst.db, err = sst.createDatabase(sst.config.Name)
	}
	if err != nil {
		return nil, err
	}

	exists, err := sst.db.GraphExists(context.TODO(), sst.name)
	if err != nil {
		return nil, err
	}
	if exists {
		sst.graph, err = sst.db.Graph(context.TODO(), sst.name)
		if err != nil {
			return nil, errors.Wrapf(err, "sst: failed to open graph: %v", sst.name)
		}
	} else {
		sst.graph, err = sst.db.CreateGraph(context.TODO(), sst.name, &arango.CreateGraphOptions{
			OrphanVertexCollections: []string{"Disconnected"},
			EdgeDefinitions: []arango.EdgeDefinition{
				{Collection: "Near", From: sst.config.NodeCollections, To: sst.config.NodeCollections},
				{Collection: "Follows", From: sst.config.NodeCollections, To: sst.config.NodeCollections},
				{Collection: "Contains", From: sst.config.NodeCollections, To: sst.config.NodeCollections},
				{Collection: "Expresses", From: sst.config.NodeCollections, To: sst.config.NodeCollections},
			},
		})
		if err != nil {
			return nil, errors.Wrapf(err, "sst: failed to create graph: %v", sst.name)
		}
	}

	sst.nodes = make(map[string]arango.Collection)
	for _, kind := range sst.config.NodeCollections {
		sst.nodes[kind], err = sst.graph.VertexCollection(context.TODO(), kind)
		if err != nil {
			return nil, errors.Wrapf(err, "sst: failed to create %v vertex collection", kind)
		}
	}

	sst.near, _, err = sst.graph.EdgeCollection(nil, "Near")
	if err != nil {
		return nil, errors.Wrap(err, "sst: failed to create Near vertex collection")
	}
	sst.follows, _, err = sst.graph.EdgeCollection(nil, "Follows")
	if err != nil {
		return nil, errors.Wrap(err, "sst: failed to create Follows vertex collection")
	}
	sst.contains, _, err = sst.graph.EdgeCollection(nil, "Contains")
	if err != nil {
		return nil, errors.Wrap(err, "sst: failed to create Contains vertex collection")
	}
	sst.expresses, _, err = sst.graph.EdgeCollection(nil, "Expresses")
	if err != nil {
		return nil, errors.Wrap(err, "sst: failed to create Expresses vertex collection")
	}

	sst.prevEvents = []*Node{startEvent}

	return sst, nil
}

// ToDocumentKey replaces disallowed characters in key names with '_'.
func ToDocumentKey(s string) string {
	return keyRegex.ReplaceAllString(s, "_")
}
