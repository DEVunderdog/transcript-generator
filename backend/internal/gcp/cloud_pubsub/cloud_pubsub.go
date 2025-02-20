package cloud_pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type CloudPubSubClient struct {
	client  pubsub.Client
	topicID string
}

func NewCloudPubSubClient(ctx context.Context, credPath, topicID, projectID string) (*CloudPubSubClient, error) {
	pubSubClient, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(credPath))
	if err != nil {
		return nil, err
	}

	return &CloudPubSubClient{
		client:  *pubSubClient,
		topicID: topicID,
	}, nil
}

func (cps *CloudPubSubClient) Close() error {
	return cps.client.Close()
}
