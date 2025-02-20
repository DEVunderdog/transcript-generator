package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type StorageClient struct {
	StorageClient storage.Client
	BucketName    string
}

func NewStorageClient(ctx context.Context, credPath, bucketName string) (*StorageClient, error) {
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsFile(credPath))
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

