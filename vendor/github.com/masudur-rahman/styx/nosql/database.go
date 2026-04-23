package nosql

import "context"

type Engine interface {
	Collection(name string) Engine

	ID(id string) Engine

	FindOne(ctx context.Context, document interface{}, filter ...interface{}) (bool, error)
	FindMany(ctx context.Context, documents interface{}, filter interface{}) error

	InsertOne(ctx context.Context, document interface{}) (id string, err error)
	InsertMany(ctx context.Context, documents []interface{}) ([]string, error)

	UpdateOne(ctx context.Context, document interface{}) error

	DeleteOne(ctx context.Context, filter ...interface{}) error

	Query(ctx context.Context, query string, bindParams map[string]interface{}) (interface{}, error)
}
