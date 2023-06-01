package testings

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dbConfig struct {
	PgHost     string
	PgPort     int
	PgUser     string
	PgPassword string
	PgDatabase string
}

// GetSeededPGDB returns a connection to seeded postgres DB for testing, DDLs are executed externally before this step
func GetSeededPGDB(seedSQL string) (*pgxpool.Pool, error) {
	dbConn, err := open(dbConfig{
		PgHost:     "localhost",
		PgPort:     5440,
		PgUser:     "user",
		PgPassword: "password",
		PgDatabase: "bench",
	})
	if err != nil {
		return nil, fmt.Errorf("opening db: %v", err)
	}

	if err := seed(dbConn, seedSQL); err != nil {
		return nil, fmt.Errorf("seeding: %v", err)
	}

	return dbConn, nil
}

func open(cfg dbConfig) (*pgxpool.Pool, error) {
	q := make(url.Values)
	q.Set("sslmode", "disable")
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.PgUser, cfg.PgPassword),
		Host:     fmt.Sprintf("%v:%v", cfg.PgHost, cfg.PgPort),
		Path:     cfg.PgDatabase,
		RawQuery: q.Encode(),
	}
	config, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("parsing config: %v", err)
	}
	config.MaxConns = 5

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func seed(db *pgxpool.Pool, seedSQL string) error {
	if seedSQL == "" {
		return nil
	}
	_, err := db.Exec(context.Background(), seedSQL)
	return err
}
