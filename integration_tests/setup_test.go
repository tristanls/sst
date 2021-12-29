package integration_tests

import (
	"context"
	"os"
	"testing"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/tristanls/sst"
)

var (
	integrationTestDBName = "integration_test__"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func arangodb(t *testing.T) arango.Database {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		t.Fatalf("integration_test: failed to create ArangoDB connection: %v", err)
	}
	client, err := arango.NewClient(arango.ClientConfig{
		Connection:     conn,
		Authentication: arango.BasicAuthentication("root", ""),
	})
	if err != nil {
		t.Fatalf("integration_test: failed to create ArangoDB client: %v", err)
	}
	exists, err := client.DatabaseExists(context.TODO(), integrationTestDBName)
	if err != nil {
		t.Fatalf("integration_test: failed to check database existence: %v", err)
	}
	if exists {
		db, err := client.Database(context.TODO(), integrationTestDBName)
		if err != nil {
			t.Fatalf("integration_test: failed to retrieve database: %v", err)
		}
		err = db.Remove(context.TODO())
		if err != nil {
			t.Fatalf("integration_test: failed to remove database: %v", err)
		}
	}
	db, err := client.CreateDatabase(context.TODO(), integrationTestDBName, nil)
	if err != nil {
		t.Fatalf("integration_test: failed to create database: %v", err)
	}
	return db
}

func st(t *testing.T) *sst.SST {
	config := &sst.Config{
		Associations: map[string]*sst.Association{
			"near":      {Key: "near", SemanticType: sst.Near, Fwd: "is near", Bwd: "is near", Nfwd: "is not near", Nbwd: "is not near"},
			"follows":   {Key: "follows", SemanticType: sst.Follows, Fwd: "follows", Bwd: "precedes", Nfwd: "does not follow", Nbwd: "does not precede"},
			"contains":  {Key: "contains", SemanticType: sst.Contains, Fwd: "contains", Bwd: "constitutes", Nfwd: "does not contain", Nbwd: "does not constitute"},
			"expresses": {Key: "expresses", SemanticType: sst.Expresses, Fwd: "expresses", Bwd: "describes", Nfwd: "does not express", Nbwd: "does not describe"},
		},
		Name:            integrationTestDBName,
		NodeCollections: []string{"Node"},
		Password:        "",
		URL:             "http://localhost:8529",
		Username:        "root",
	}
	s, err := sst.NewSST(config)
	if err != nil {
		t.Fatalf("integration_test: failed to create SST: %v", config)
	}
	return s
}
