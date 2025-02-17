from google.cloud import pubsub_v1
from gcp.cloud_pubsub import CloudPubSub
from gcp.cloud_storage import CloudStorage
from audio_model.audio_model import ASRModel
from custom_proto.message_pb2 import TopicMessage
from logger import logger


class Service:
    def __init__(self, project_id: str, bucket_name: str, subscription_id: str):
        self.cloudPubSub = CloudPubSub(
            project_id=project_id, subscription_id=subscription_id
        )
        self.cloudStorage = CloudStorage(project_id=project_id, bucket_name=bucket_name)
        self.asrModel = ASRModel()

    def custom_callback(self, message: pubsub_v1.subscriber.message.Message):
        try:
            topic_message = TopicMessage()
            topic_message.ParseFromString(message.data)
            message.ack()

            object_key = topic_message.object_key
            objects_list = [object_key]

            file, file_name = self.cloudStorage.download_audio_files(objects_list)

            resample_file = self.asrModel.resample_file(file=file, file_name=file_name)

            wav_model, processor = self.asrModel.instantiate_model()

            transcript = self.asrModel.asr_transcript(
                processor=processor, model=wav_model, resampled_path=resample_file
            )

            logger.info(f"transcipt: {transcript}")

        except Exception as e:
            logger.error(f"error parsing the message {str(e)}")
            message.nack()

    def run_service(self):
        self.cloudPubSub.start_listening(callback=self.custom_callback)
