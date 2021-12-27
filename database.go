package sst

import (
	"context"

	arango "github.com/arangodb/go-driver"
	"github.com/pkg/errors"
)

// CreateDatabase creates a new ArangoDB database if it does not exist.
func (s *SST) CreateDatabase(name string) (arango.Database, error) {
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

// OpenDatabase opens an ArangoDB database if it exists.
func (s *SST) OpenDatabase(name string) (arango.Database, error) {
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
