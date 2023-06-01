package rules

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mrkagelui/pg-mongo-benchmarks/testings"
)

var (
	celEnv  *cel.Env
	pgPool  *pgxpool.Pool
	mongoDB *mongo.Database
)

func TestMain(m *testing.M) {
	env, err := cel.NewEnv(
		ext.NativeTypes(reflect.TypeOf(Txn{})),
		cel.Variable("txn", cel.ObjectType("rules.Txn")),
	)
	if err != nil {
		log.Fatal(err)
	}
	celEnv = env

	pgDB, err := testings.GetSeededPGDB("")
	if err != nil {
		log.Fatal(err)
	}
	defer pgDB.Close()
	pgPool = pgDB

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := testings.GetMongoDB(ctx, "mongodb://root:password123@localhost:27017/?replicaSet=replicaset&directConnection=true")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := c.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	mongoDB = c.Database("bench")

	m.Run()
}

func Test_pgStore_Create(t *testing.T) {
	tests := []struct {
		name   string
		rule   NewRule
		assert testings.ErrCheck
	}{
		{
			name: "normal",
			rule: NewRule{
				Name:       "tt1",
				UserID:     "12345",
				Definition: "txn.Amount >= 1000000.0",
				IsActive:   true,
				Included:   []string{"abc"},
				Excluded:   []string{"cde"},
			},
			assert: testings.Ok,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := pgPool.Begin(context.Background())
			testings.Ok(t, err)
			defer tx.Rollback(context.Background())

			s := &PgStore{
				env:  celEnv,
				pool: tx,
			}
			got, err := s.Create(context.Background(), tt.rule)
			tt.assert(t, err)
			gotNewRule, err := s.getByID(context.Background(), got)
			testings.Ok(t, err)
			if !reflect.DeepEqual(gotNewRule, tt.rule) {
				t.Errorf("got %v in DB, want %v", gotNewRule, tt.rule)
			}
		})
	}
}

func (s *PgStore) getByID(ctx context.Context, id string) (NewRule, error) {
	rows, err := s.pool.Query(ctx, `
	SELECT name, user_id, definition, is_active, included_entities, excluded_entities FROM rules WHERE id = $1`, id)
	if err != nil {
		return NewRule{}, fmt.Errorf("query: %v", err)
	}
	nr, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByPos[NewRule])
	if err != nil {
		return NewRule{}, fmt.Errorf("collecting: %v", err)
	}

	return *nr, nil
}

func Test_mongoStore_Create(t *testing.T) {
	tests := []struct {
		name   string
		rule   NewRule
		assert testings.ErrCheck
	}{
		{
			name: "normal",
			rule: NewRule{
				Name:       "tt1",
				UserID:     "12345",
				Definition: "txn.Amount >= 1000000.0",
				IsActive:   true,
				Included:   []string{"abc"},
				Excluded:   []string{"cde"},
			},
			assert: testings.Ok,
		},
	}
	for _, tt := range tests {
		session, err := mongoDB.Client().StartSession()
		testings.Ok(t, err)

		testings.Ok(t, mongo.WithSession(context.Background(), session, func(ctx mongo.SessionContext) error {
			t.Run(tt.name, func(t *testing.T) {
				testings.Ok(t, session.StartTransaction())
				m := MongoStore{
					env: celEnv,
					db:  mongoDB,
				}
				got, err := m.Create(ctx, tt.rule)
				tt.assert(t, err)
				gotNewRule, err := m.getByID(ctx, got)
				testings.Ok(t, err)
				if !reflect.DeepEqual(gotNewRule, tt.rule) {
					t.Errorf("got %v in DB, want %v", gotNewRule, tt.rule)
				}

				//testings.Ok(t, session.AbortTransaction(context.Background()))
				testings.Ok(t, session.CommitTransaction(context.Background()))
			})
			return nil
		}))
	}
}

func (m *MongoStore) getByID(ctx context.Context, id any) (NewRule, error) {
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

	var nr newRule
	if err := m.db.Collection("rules").FindOne(ctx, bson.D{{"_id", id}}).Decode(&nr); err != nil {
		return NewRule{}, err
	}

	return NewRule{
		Name:       nr.Name,
		UserID:     nr.UserID,
		Definition: nr.Definition,
		IsActive:   nr.IsActive,
		Included:   nr.IncludedEntities,
		Excluded:   nr.ExcludedEntities,
	}, nil
}
