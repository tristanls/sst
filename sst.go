// Package sst provides facilities for modeling Semantic Spacetime
// in an ArangoDB.
package sst

import (
	"context"
	"fmt"
	"regexp"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/pkg/errors"
)

var (
	databaseAlreadyExists = errors.New("sst: database already exists")
	databaseDoesNotExist  = errors.New("sst: database does not exist")
	startEvent            = &Node{Key: "start"}
)

type Config struct {
	Name string
	// NodeCollections are the names of node collections to instantiate for this SST
	NodeCollections []string
	Password        string
	URL             string
	Username        string
}

type SST struct {
	db     arango.Database
	client arango.Client
	config *Config
	conn   arango.Connection
	graph  arango.Graph
	name   string

	nodes map[string]arango.Collection

	follows   arango.Collection
	contains  arango.Collection
	expresses arango.Collection
	near      arango.Collection

	prevEvents []*Node
}

type SemanticType int

const (
	Near SemanticType = iota
	Follows
	Contains
	Expresses
)

func (a SemanticType) String() string {
	switch a {
	case Near:
		return "Near"
	case Follows:
		return "Follows"
	case -Follows:
		return "Precedes"
	case Contains:
		return "Contains"
	case -Contains:
		return "Constitutes"
	case Expresses:
		return "Expresses"
	case -Expresses:
		return "Describes"
	}
	return "unknown"
}

// Association stores invariant relationship data as lookup tables to
// reduce database storage.
type Association struct {
	Key          string       `json:"_key"`
	SemanticType SemanticType `json:"stype"`
	Fwd          string       `json:"fwd"`
	Bwd          string       `json:"bwd"`
	Nfwd         string       `json:"nfwd"`
	Nbwd         string       `json:"nbwd"`
}

var (
	associations = map[string]*Association{
		"contains":        {"contains", Contains, "contains", "belongs to or is part of", "does not contain", "is not part of"},
		"generalizes":     {"generalizes", Contains, "generalizes", "is a special case of", "is not a generalization of", "is not a special case of"},
		"part_of":         {"part_of", -Contains, "is part of", "incorporates", "is not part of", "doesn't incorporate"},
		"has_role":        {"has_role", Expresses, "has the role of", "is a role fulfilled by", "has no role", "is not a role fulfilled by"},
		"originates_from": {"originates_from", Follows, "originates from", "is the source/origin of", "does not originate from", "is not the source/origin of"},
		"expresses":       {"expresses", Expresses, "expresses an attribute", "is an attribute of", "has no attribute", "is not an attribute of"},
		"promises":        {"promises", Expresses, "promises/intends", "is intended/promised by", "rejects/promises to not", "is rejected by"},
		"has_name":        {"has_name", Expresses, "has proper name", "is the proper name of", "is not named", "isn't the proper name of"},
		"follows_from":    {"follows_from", Follows, "follows on from", "is followed by", "does not follow", "does not precede"},
		"uses":            {"uses", Follows, "uses", "is used by", "does not use", "is not used by"},
		"caused_by":       {"caused_by", Follows, "caused by", "may cause", "was not caused by", "probably didn't cause"},
		"derives_from":    {"derives_from", Follows, "derives from", "leads to", "does not derive from", "does not lead to"},
		"depends":         {"depends", Follows, "may depends on", "may determine", "doesn't depend on", "doesn't determine"},
		"next":            {"next", -Follows, "comes before", "comes after", "is not before", "is not after"},
		"then":            {"then", -Follows, "then", "previously", "but not", "didn't follow"},
		"leads_to":        {"leads_to", -Follows, "leads to", "doesn't imply", "doesn't reach", "doesn't precede"},
		"precedes":        {"precedes", -Follows, "precedes", "follows", "doesn't precede", "doesn't follow"},
		"related":         {"related", Near, "may be related to", "may be related to", "likely unrelated to", "likely unrelated to"},
		"alias":           {"alias", Near, "also known as", "also known as", "not known as", "not known as"},
		"is_like":         {"is_like", Near, "is similar to", "is similar to", "is unlike", "is unlike"},
		"connected":       {"connected", Near, "is connected to", "is connected to", "is not connected to", "is not connected to"},
		"coactive":        {"coactive", Near, "occurred together with", "occurred together with", "never appears with", "never appears with"},
	}
)

var (
	keyRegex = regexp.MustCompile(`[^a-zA-Z0-9_:.@()+,=;$!*'%-]`)
)

// Creates new Semantic Spacetime model backed by ArangoDB
func NewSST(config *Config) (*SST, error) {
	sst := &SST{
		config: config,
		name:   "semantic_spacetime",
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

// CreateAssociation creates a new association
func CreateAssociation(a *Association) error {
	a.Key = toDocumentKey(a.Key)
	existing := associations[a.Key]
	if existing == nil {
		associations[a.Key] = a
		return nil
	}
	if existing == a {
		return nil
	}
	return errors.New(fmt.Sprintf("sst: failed to create association %v due to existing association %v", a, existing))
}

// MustCreateAssociation creates a new association, panics on error
func MustCreateAssociation(a *Association) {
	err := CreateAssociation(a)
	if err != nil {
		panic(err)
	}
}

func toDocumentKey(s string) string {
	return keyRegex.ReplaceAllString(s, "_")
}
