from google.cloud import pubsub_v1
from transcript_service.config import settings

timeout = 5.0

subscriber = pubsub_v1.SubscriberClient()
subscription_path = subscriber.subscription_path(project=settings.project_id, subscription=settings.subscription_id)

def callback(message: pubsub_v1.subscriber.message.Message) -> None:
    print(f"Received {message}")
    message.ack()

streaming_pull_future = subscriber.subscribe(subscription=subscription_path, callback=callback)
print(f"Listening for message on {subscription_path}..\n")

with subscriber:
    try:
        streaming_pull_future.result(timeout=timeout)
    except Exception as e:
        print(
            f"listening for message on {subscription_path} threw an exception: {e}"
        )
        streaming_pull_future.cancel()
        streaming_pull_future.result()