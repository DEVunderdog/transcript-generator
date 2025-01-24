package database

import (
	"context"
	"errors"

	custom_errors "github.com/DEVunderdog/transcript-generator-backend/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UpdateFileMetadataTxParams struct {
	ID         int32
	UserID     int32
	ObjectKey  pgtype.Text
	UpdatedAt  pgtype.Timestamptz
	FileStatus string
}

func (store *SQLStore) CreateEmptyFileTx(ctx context.Context, arg CreateEmptyFileParams) (*FileRegistry, error) {

	var file FileRegistry

	err := store.execTx(ctx, func(q *Queries) error {

		var err error

		file, err = q.CreateEmptyFile(ctx, CreateEmptyFileParams{
			UserID:       arg.UserID,
			FileName:     arg.FileName,
			LockStatus:   Locked,
			UploadStatus: Pending,
		})

		if err != nil {
			if ErrorCode(err) == UniqueViolation {
				return custom_errors.ErrDuplicateData
			}
			return err
		}

		return nil

	})

	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (store *SQLStore) UpdateMetadataFileTx(ctx context.Context, arg UpdateFileMetadataTxParams) (*FileRegistry, error) {

	var file FileRegistry

	err := store.execTx(ctx, func(q *Queries) error {

		var err error

		fileData, err := q.GetFileByID(ctx, GetFileByIDParams{
			ID:     arg.ID,
			UserID: arg.UserID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return custom_errors.ErrNoRecordFound
			}
			return err
		}

		if !fileData.LockStatus || fileData.UpdatedAt != arg.UpdatedAt || fileData.UploadStatus != Pending {
			return custom_errors.ErrResourceConflict
		}

		file, err = q.UpdateFileMetadata(ctx, UpdateFileMetadataParams{
			ObjectKey:    arg.ObjectKey,
			UploadStatus: arg.FileStatus,
			LockStatus:   Unlocked,
			ID:           arg.ID,
			UserID:       arg.UserID,
		})

		if err != nil {
			return err
		}

		return nil

	})

	if err != nil {
		return nil, err
	}

	return &file, nil

}

func (store *SQLStore) UpdateFileNameTx(ctx context.Context, userID int32, oldFilename, newFilename string) (*FileRegistry, error) {
	var file FileRegistry

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		fileData, err := q.GetFileByName(ctx, GetFileByNameParams{
			FileName: oldFilename,
			UserID:   userID,
		})

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return custom_errors.ErrNoRecordFound
			}
			return err
		}

		if fileData.LockStatus {
			return custom_errors.ErrResourceLocked
		}

		if fileData.UploadStatus != Success {
			return custom_errors.ErrUploadIssue
		}

		file, err = q.UpdateFileName(ctx, UpdateFileNameParams{
			NewFileName: newFilename,
			ID:          fileData.ID,
		})

		if err != nil {
			if ErrorCode(err) == UniqueViolation {
				return custom_errors.ErrDuplicateData
			}

			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (store *SQLStore) LockFileTx(ctx context.Context, userID int32, filename string) (*FileRegistry, error) {
	var file FileRegistry

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		fileData, err := q.GetFileByName(ctx, GetFileByNameParams{
			FileName: filename,
			UserID:   userID,
		})

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return custom_errors.ErrNoRecordFound
			}

			return err
		}

		if fileData.LockStatus {
			return custom_errors.ErrResourceLocked
		}

		if fileData.UploadStatus != Success {
			return custom_errors.ErrUploadIssue
		}

		file, err = q.UnlockAndLockFile(ctx, UnlockAndLockFileParams{
			LockStatus: Locked,
			ID:         fileData.ID,
			UserID:     userID,
			Status:     Pending,
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (store *SQLStore) DeleteFileTx(ctx context.Context, userId, id int32, updatedAt pgtype.Timestamptz) error {

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		fileData, err := q.GetFileByID(ctx, GetFileByIDParams{
			ID:     id,
			UserID: userId,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return custom_errors.ErrNoRecordFound
			}
			return err
		}

		if !fileData.LockStatus || fileData.UpdatedAt != updatedAt {
			return custom_errors.ErrResourceConflict
		}

		if fileData.UploadStatus != Pending {
			return custom_errors.ErrUploadIssue
		}

		err = q.DeleteFile(ctx, DeleteFileParams{
			ID:     id,
			UserID: userId,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
