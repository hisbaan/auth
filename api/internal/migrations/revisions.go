package migrations

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ariga.io/atlas/sql/migrate"
	"github.com/lib/pq"
)

const (
	revisionSchema = "public"
	revisionTable  = "atlas_schema_revisions"

	createRevisionTableSQL = `
		CREATE TABLE IF NOT EXISTS public.atlas_schema_revisions (
			version varchar(255) PRIMARY KEY,
			description varchar(255) NOT NULL DEFAULT '',
			type bigint NOT NULL DEFAULT 0,
			applied integer NOT NULL DEFAULT 0,
			total integer NOT NULL DEFAULT 0,
			executed_at timestamptz NOT NULL,
			execution_time bigint NOT NULL DEFAULT 0,
			error text NOT NULL DEFAULT '',
			error_stmt text NOT NULL DEFAULT '',
			hash text NOT NULL DEFAULT '',
			partial_hashes text[] NOT NULL DEFAULT '{}',
			operator_version text NOT NULL DEFAULT ''
		)
	`

	selectRevisionsSQL = `
		SELECT version, description, type, applied, total, executed_at, execution_time,
		       error, error_stmt, hash, partial_hashes, operator_version
		FROM public.atlas_schema_revisions
		ORDER BY version
	`

	selectRevisionSQL = `
		SELECT version, description, type, applied, total, executed_at, execution_time,
		       error, error_stmt, hash, partial_hashes, operator_version
		FROM public.atlas_schema_revisions
		WHERE version = $1
	`

	writeRevisionSQL = `
		INSERT INTO public.atlas_schema_revisions
		(version, description, type, applied, total, executed_at, execution_time,
		 error, error_stmt, hash, partial_hashes, operator_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (version) DO UPDATE SET
			description = EXCLUDED.description,
			type = EXCLUDED.type,
			applied = EXCLUDED.applied,
			total = EXCLUDED.total,
			executed_at = EXCLUDED.executed_at,
			execution_time = EXCLUDED.execution_time,
			error = EXCLUDED.error,
			error_stmt = EXCLUDED.error_stmt,
			hash = EXCLUDED.hash,
			partial_hashes = EXCLUDED.partial_hashes,
			operator_version = EXCLUDED.operator_version
	`

	deleteRevisionSQL = `DELETE FROM public.atlas_schema_revisions WHERE version = $1`
)

// RevisionStore stores Atlas migration revisions in Postgres.
//
// The Atlas CLI manages this table internally. The Go executor expects callers
// to provide revision storage, so this type implements Atlas' default table.
type RevisionStore struct {
	db *sql.DB
}

func NewRevisionStore(ctx context.Context, db *sql.DB) (*RevisionStore, error) {
	store := &RevisionStore{db: db}
	if err := store.ensureTable(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

func (*RevisionStore) Ident() *migrate.TableIdent {
	return &migrate.TableIdent{Name: revisionTable, Schema: revisionSchema}
}

func (s *RevisionStore) ReadRevisions(ctx context.Context) ([]*migrate.Revision, error) {
	rows, err := s.db.QueryContext(ctx, selectRevisionsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revisions []*migrate.Revision
	for rows.Next() {
		rev, err := scanRevision(rows)
		if err != nil {
			return nil, err
		}
		revisions = append(revisions, rev)
	}
	return revisions, rows.Err()
}

func (s *RevisionStore) ReadRevision(ctx context.Context, version string) (*migrate.Revision, error) {
	row := s.db.QueryRowContext(ctx, selectRevisionSQL, version)
	rev, err := scanRevision(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, migrate.ErrRevisionNotExist
	}
	return rev, err
}

func (s *RevisionStore) WriteRevision(ctx context.Context, rev *migrate.Revision) error {
	partialHashes := rev.PartialHashes
	if partialHashes == nil {
		partialHashes = []string{}
	}

	_, err := s.db.ExecContext(ctx, writeRevisionSQL,
		rev.Version,
		rev.Description,
		uint(rev.Type),
		rev.Applied,
		rev.Total,
		rev.ExecutedAt,
		rev.ExecutionTime.Nanoseconds(),
		rev.Error,
		rev.ErrorStmt,
		rev.Hash,
		pq.Array(partialHashes),
		rev.OperatorVersion,
	)
	return err
}

func (s *RevisionStore) DeleteRevision(ctx context.Context, version string) error {
	_, err := s.db.ExecContext(ctx, deleteRevisionSQL, version)
	return err
}

func (s *RevisionStore) ensureTable(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, createRevisionTableSQL)
	return err
}

type revisionScanner interface {
	Scan(dest ...any) error
}

func scanRevision(scanner revisionScanner) (*migrate.Revision, error) {
	var (
		rev           migrate.Revision
		typ           int64
		executionTime int64
	)
	if err := scanner.Scan(
		&rev.Version,
		&rev.Description,
		&typ,
		&rev.Applied,
		&rev.Total,
		&rev.ExecutedAt,
		&executionTime,
		&rev.Error,
		&rev.ErrorStmt,
		&rev.Hash,
		pq.Array(&rev.PartialHashes),
		&rev.OperatorVersion,
	); err != nil {
		return nil, err
	}
	rev.Type = migrate.RevisionType(typ)
	rev.ExecutionTime = time.Duration(executionTime)
	return &rev, nil
}
