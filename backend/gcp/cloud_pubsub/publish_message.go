package cloud_pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/DEVunderdog/transcript-generator-backend/proto/pb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (cps *CloudPubSubClient) PublishMessage(ctx context.Context, email, objectKey string, userID int) error {
	msg := &pb.TopicMessage{
		ObjectKey: objectKey,
		UserId:    int64(userID),
		UserEmail: email,
	}

	topic := cps.client.Topic(cps.topicID)

	topicConfig, err := topic.Config(ctx)
	if err != nil {
		return fmt.Errorf("error while configuring topic.Config: %w", err)
	}

	var data []byte
	switch topicConfig.SchemaSettings.Encoding {
	case pubsub.EncodingJSON:
		data, err = protojson.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error while encoding json case protojson.Marshal: %w", err)
		}
	case pubsub.EncodingBinary:
		data, err = proto.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error while encoding binary case proto.Marshal: %w", err)
		}
	default:
		return fmt.Errorf("unknown encoding: %v", topicConfig.SchemaSettings.Encoding)
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("error while publishing the message to topic: %w", err)
	}

	return nil
}
