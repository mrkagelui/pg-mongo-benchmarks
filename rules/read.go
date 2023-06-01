package rules

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/bson"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// Read queries the DB and returns all applicable rules
func (s *PgStore) Read(ctx context.Context, userID string, entityID string) ([]ExecutableRule, error) {
	q := `
SELECT id, name, compiled
FROM rules
WHERE user_id = $1
  AND is_active
  AND (included_entities IS NULL OR included_entities @> ARRAY [$2])
  AND NOT rules.excluded_entities @> ARRAY [$2]`

	rows, err := s.pool.Query(ctx, q, userID, entityID)
	if err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}

	type temp struct {
		ID       string
		Name     string
		Compiled []byte
	}

	res, err := pgx.CollectRows(rows, pgx.RowToStructByPos[temp])
	if err != nil {
		return nil, fmt.Errorf("collecting: %v", err)
	}

	result := make([]ExecutableRule, len(res))
	for i, r := range res {
		exp := &expr.CheckedExpr{}
		if err := proto.Unmarshal(r.Compiled, exp); err != nil {
			return nil, fmt.Errorf("unmarshal: %v", err)
		}

		result[i] = ExecutableRule{
			ID:   r.ID,
			Name: r.Name,
			Ast:  cel.CheckedExprToAst(exp),
		}
	}

	return result, nil
}

// Read queries the DB and returns all applicable rules
func (m *MongoStore) Read(ctx context.Context, userID string, entityID string) ([]ExecutableRule, error) {
	_ = bson.D{
		{
			Key: "$and",
			Value: bson.A{
				bson.D{
					{"user_id", userID},
				},
				bson.D{
					{"is_active", true},
				},
				bson.D{
					{"$or", bson.A{
						bson.D{
							{"included_entities", bson.D{{"$exists", false}}},
						},
						bson.D{
							{"included_entities", entityID},
						},
					},
					}},
				bson.D{
					{"$not", bson.D{
						{"excluded_entities", entityID},
					},
					},
				},
			},
		},
	}

	return nil, nil
}
