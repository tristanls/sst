package sst

import (
	"context"

	"github.com/arangodb/go-driver"
)

// Query executes the designated ArangoDB query
func (s *SST) Query(ctx context.Context, query string, vars map[string]interface{}) (driver.Cursor, error) {
	return s.db.Query(ctx, query, vars)
}
