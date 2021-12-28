package sst

import (
	"context"

	arango "github.com/arangodb/go-driver"
	"github.com/pkg/errors"
)

var (
	databaseAlreadyExists = errors.New("sst: database already exists")
	databaseDoesNotExist  = errors.New("sst: database does not exist")
)

// createDatabase creates a new ArangoDB database if it does not exist.
func (s *SST) createDatabase(name string) (arango.Database, error) {
	ctx := context.Background()
	exists, err := s.client.DatabaseExists(ctx, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, databaseAlreadyExists
	}
	db, err := s.client.CreateDatabase(ctx, name, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "sst: failed to create database: %v", name)
	}
	return db, nil
}

// openDatabase opens an ArangoDB database if it exists.
func (s *SST) openDatabase(name string) (arango.Database, error) {
	ctx := context.Background()
	exists, err := s.client.DatabaseExists(ctx, name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, databaseDoesNotExist
	}
	db, err := s.client.Database(ctx, name)
	if err != nil {
		return nil, err
	}
	return db, nil
}
