package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	CreateEmptyFileTx(ctx context.Context, arg CreateEmptyFileParams) (*FileRegistry, error)
	UpdateMetadataFileTx(ctx context.Context, arg UpdateFileMetadataTxParams) (*FileRegistry, error)
	UpdateFileNameTx(ctx context.Context, userID int32, oldFilename, newFilename string) (*FileRegistry, error)
	LockFileTx(ctx context.Context, userID int32, filename string) (*FileRegistry, error)
	DeleteFileTx(ctx context.Context, userId, id int32, updatedAt pgtype.Timestamptz) error
}

type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}
