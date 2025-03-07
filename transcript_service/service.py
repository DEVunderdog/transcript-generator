import sys
from pathlib import Path

from audio_model.audio_model import ASRModel
from constants import constants
from custom_proto.message_pb2 import TopicMessage
from gcp.cloud_pubsub import CloudPubSub
from gcp.cloud_storage import CloudStorage
from google.cloud import pubsub_v1
from logger import logger
from mail.transcript_email import TranscriptEmail
from pdf.generate_pdf import PdfProcessor


class Service:
    def __init__(
        self,
        project_id: str,
        bucket_name: str,
        subscription_id: str,
        sender_email: str,
        sender_password: str,
    ):

        self.cloudPubSub = CloudPubSub(
            project_id=project_id, subscription_id=subscription_id
        )
        self.cloudStorage = CloudStorage(project_id=project_id, bucket_name=bucket_name)
        self.asrModel = ASRModel()
        self.pdfProcessor = PdfProcessor()
        self.emailProcessor = TranscriptEmail(
            sender_email=sender_email,
            sender_password=sender_password,
        )


    def custom_callback(self, message: pubsub_v1.subscriber.message.Message):
        try:
            topic_message = TopicMessage()
            topic_message.ParseFromString(message.data)
            message.ack()

            object_key = topic_message.object_key
            objects_list = [object_key]

            user_email = topic_message.user_email

            file, file_name = self.cloudStorage.download_audio_files(objects_list)

            resample_file = self.asrModel.resample_file(file=file, file_name=file_name)

            model, processor = self.asrModel.instantiate_model()

            transcript = self.asrModel.generate_transcript(
                model=model, processor=processor, file=resample_file
            )

            self.pdfProcessor.generate_pdf(content=transcript)
            self.emailProcessor.send_email(recipient_email=user_email)

            logger.info("mail sent successfully")

        except Exception as e:
            logger.error(f"error parsing the message {str(e)}")
            message.nack()

    def run_service(self):
        self.cloudPubSub.start_listening(callback=self.custom_callback)

    def cleanup(self):
        temp_dir = Path(constants.temp_dir)

        for item in temp_dir.iterdir():
            if item.is_file():
                item.unlink()
            elif item.is_dir():
                for sub_item in item.iterdir():
                    if sub_item.is_file():
                        sub_item.unlink()

    def signal_handler(self, sig, frame):
        logger.info("termination signal received cleaning up...")
        self.cleanup()
        sys.exit(0)
