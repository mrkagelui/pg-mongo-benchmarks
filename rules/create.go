package rules

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/cel-go/cel"
	"google.golang.org/protobuf/proto"
)

func compile(env *cel.Env, def string) ([]byte, error) {
	ast, iss := env.Compile(def)
	if err := iss.Err(); err != nil {
		return nil, fmt.Errorf("compiling: %v", err)
	}
	if ot := ast.OutputType(); !reflect.DeepEqual(ot, cel.BoolType) {
		return nil, fmt.Errorf("got %v, but want bool type", ot)
	}
	exp, err := cel.AstToCheckedExpr(ast)
	if err != nil {
		return nil, fmt.Errorf("to checked exp: %v", err)
	}
	b, err := proto.Marshal(exp)
	if err != nil {
		return nil, fmt.Errorf("proto marshal: %v", err)
	}

	return b, nil
}

// Create inserts one NewRule into a postgres DB
func (s *PgStore) Create(ctx context.Context, rule NewRule) (string, error) {
	b, err := compile(s.env, rule.Definition)
	if err != nil {
		return "", fmt.Errorf("compile: %v", err)
	}

	var id string
	if err := s.pool.QueryRow(ctx, `INSERT INTO rules
(name, user_id, definition, compiled, is_active, included_entities, excluded_entities)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id`,
		rule.Name,
		rule.UserID,
		rule.Definition,
		b,
		rule.IsActive,
		rule.Included,
		rule.Excluded).Scan(&id); err != nil {
		return "", fmt.Errorf("inserting: %v", err)
	}
	return id, err
}

// Create inserts one NewRule into a mongo DB
func (m *MongoStore) Create(ctx context.Context, rule NewRule) (any, error) {
	b, err := compile(m.env, rule.Definition)
	if err != nil {
		return 0, fmt.Errorf("compile: %v", err)
	}

	type newRule struct {
		CreatedAt        time.Time `bson:"created_at"`
		UpdatedAt        time.Time `bson:"updated_at"`
		Name             string    `bson:"name"`
		UserID           string    `bson:"user_id"`
		Definition       string    `bson:"definition"`
		Compiled         []byte    `bson:"compiled"`
		IsActive         bool      `bson:"is_active"`
		IncludedEntities []string  `bson:"included_entities"`
		ExcludedEntities []string  `bson:"excluded_entities"`
	}

	r, err := m.db.Collection("rules").InsertOne(ctx, newRule{
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Name:             rule.Name,
		UserID:           rule.UserID,
		Definition:       rule.Definition,
		Compiled:         b,
		IsActive:         rule.IsActive,
		IncludedEntities: rule.Included,
		ExcludedEntities: rule.Excluded,
	})
	if err != nil {
		return 0, fmt.Errorf("insert one: %v", err)
	}
	return r.InsertedID, nil
}
