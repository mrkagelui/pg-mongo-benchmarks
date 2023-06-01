// Package rules provides means to manage rules in persistence
package rules

import (
	"context"

	"github.com/google/cel-go/cel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.mongodb.org/mongo-driver/mongo"
)

// ExecutableRule contains all needed to evaluate a Transaction
type ExecutableRule struct {
	ID   string
	Name string
	Ast  *cel.Ast
}

// NewRule contains all needed to insert a new rule
type NewRule struct {
	Name       string
	UserID     string
	Definition string
	IsActive   bool
	Included   []string
	Excluded   []string
}

// pgxDB is something that provides DB querying access.
type pgxDB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

// PgStore holds all needed to manage rules in a postgres DB
type PgStore struct {
	env  *cel.Env
	pool pgxDB
}

// MongoStore holds all needed to manage rules
type MongoStore struct {
	env *cel.Env
	db  *mongo.Database
}

// Txn is the model for a transaction
type Txn struct {
	Name          string
	Type          string
	Currency      string
	Amount        float64
	RiskScore     int
	CustomData    CustomData
	CustomBools   map[string]bool
	CustomFloats  map[string]float64
	CustomStrings map[string]string
	Aggregates    map[string]float64
}

// CustomData is what could be sent in, data of any shape
type CustomData map[string]any
