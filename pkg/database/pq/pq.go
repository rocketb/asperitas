package db

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/rocketb/asperitas/pkg/logger"
	"github.com/rocketb/asperitas/pkg/web"
)

// lib/pg errorCodeNames
const (
	uniqueViolation = "23505"
	undefinedTable  = "42P01"
)

var (
	ErrDBNotFound        = errors.New("not found")
	ErrDBDuplicatedEntry = errors.New("duplicated entry")
	ErrUndefinedTable    = errors.New("undefined table")
)

type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

// Open used to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

// StatusCheck checks if it can successfully talk to the database.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	var tmp bool
	return db.QueryRowContext(ctx, `SELECT true`).Scan(&tmp)
}

// NamedExecContext is a helper function to execute a CUD operation
// where field replacement is necessary.
func NamedExecContext(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any) error {
	q := queryString(query, data)

	log.Infoc(ctx, 5, "database.NamedExecContext", query, q)

	ctx, span := web.AddSpan(ctx, "pkg.database.NamedExecContext", attribute.String("query", q))
	defer span.End()

	if _, err := sqlx.NamedExecContext(ctx, db, query, data); err != nil {
		if pqerr, ok := err.(*pq.Error); ok {
			switch pqerr.Code {
			case undefinedTable:
				return ErrUndefinedTable
			case uniqueViolation:
				return ErrDBDuplicatedEntry
			}
		}
		return err
	}

	return nil
}

// QuerySlice is a helper function to executing queries that
// return a collection of data to be unmarshalled into a slice.
func QuerySlice[T any](ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, dest *[]T) error {
	return namedQuerySlice(ctx, log, db, query, struct{}{}, dest)
}

// NamedQuerySlice is a helper function to executing queries that
// return a collection of data to be unmarshalled into a slice where
// field replacement is necessary.
func NamedQuerySlice[T any](ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any, dest *[]T) error {
	return namedQuerySlice(ctx, log, db, query, data, dest)
}

func namedQuerySlice[T any](ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any, dest *[]T) error {
	q := queryString(query, data)

	log.Infoc(ctx, 5, "pkg.data.pq.namedQuerySlice", "trace_id", web.GetTraceID(ctx), "query", q)

	ctx, span := web.AddSpan(ctx, "pkg.data.pq.namedQuerySlice", attribute.String("query", q))
	defer span.End()

	rows, err := sqlx.NamedQueryContext(ctx, db, query, data)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == undefinedTable {
			return ErrUndefinedTable
		}
		return err
	}
	defer rows.Close()

	var res []T
	for rows.Next() {
		v := new(T)
		if err := rows.StructScan(v); err != nil {
			return err
		}
		res = append(res, *v)
	}
	*dest = res

	return nil
}

// NamedQueryStruct is used to return a single value to be unmarshalled into a struct type
// where field replacement is necessary.
func NamedQueryStruct(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any, dest any) error {
	q := queryString(query, data)

	log.Infoc(ctx, 5, "pkg.data.pq.NamedQueryStruct", "trace_id", web.GetTraceID(ctx), "query", q)

	ctx, span := web.AddSpan(ctx, "pkg.database.query", attribute.String("query", q))
	defer span.End()

	rows, err := sqlx.NamedQueryContext(ctx, db, query, data)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == undefinedTable {
			return ErrUndefinedTable
		}
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrDBNotFound
	}

	return rows.StructScan(dest)
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args any) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("%q", v)
		case []byte:
			value = fmt.Sprintf("%q", string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}
