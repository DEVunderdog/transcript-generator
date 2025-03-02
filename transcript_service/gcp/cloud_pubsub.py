from google.cloud import pubsub_v1
from typing import Callable
from logger import logger

class CloudPubSub:
    def __init__(self, project_id: str, subscription_id: str):
        self.project_id = project_id
        self.subscription_id = subscription_id
        self.subscriber = pubsub_v1.SubscriberClient()
        self.subscription_path = self.subscriber.subscription_path(
            self.project_id, self.subscription_id
        )
        self.streaming_pull_future = None

    def start_listening(
        self, callback: Callable[[pubsub_v1.subscriber.message.Message], None]
    ):
        logger.info(f"listening message on {self.subscription_path}..\n")

        self.streaming_pull_future = self.subscriber.subscribe(
            self.subscription_path, callback=callback
        )

        with self.subscriber:
            try:
                self.streaming_pull_future.result(timeout=None)
            except KeyboardInterrupt:
                logger.info("keyboard interrupt detected within start listening")
                self.stop_listening()
            except TimeoutError:
                logger.info(
                    "timeout occurred, no messages received within the timeout period"
                )
            except Exception as e:
                logger.error(
                    f"listening for messages on {self.subscription_path} threw an exception: {str(e)}"
                )
                self.streaming_pull_future.cancel()

    def stop_listening(self):
        if self.streaming_pull_future:
            logger.info("subscription stopped")
            self.streaming_pull_future.cancel()
            try:
                self.streaming_pull_future.result()
            except Exception as e:
                logger.error(f"exception during shutdown: {str(e)}")
            finally:
                self.streaming_pull_future = None
