package storage

import (
	"context"

	"cloud.google.com/go/storage"
)

type StorageClient struct {
	StorageClient storage.Client
	BucketName    string
}

func NewStorageClient(ctx context.Context, bucketName string) (*StorageClient, error) {
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &StorageClient{
		StorageClient: *storageClient,
		BucketName: bucketName,
	}, nil
}

func (sc *StorageClient) Close() error {
	return sc.StorageClient.Close()
}

