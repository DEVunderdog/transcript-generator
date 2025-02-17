import smtplib
from email import encoders
from email.mime.base import MIMEBase
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from constants import constants


class TranscriptEmail:
    def __init__(self, sender_email, sender_password):
        self.sender_email = sender_email
        self.sender_password = sender_password
        self.subject = "transcript pdf file"
        self.body = "We have attached pdf file with transcript written in that"

    def send_email(self, recipient_email: str):
        file = constants.transcript_dir + "/transcript.pdf"
        with open(file, "rb") as attachment:
            part = MIMEBase("application", "octet-stream")
            part.set_payload(attachment.read())

        encoders.encode_base64(part)
        part.add_header("Content-Disposition", "attachment; filename= 'transcript.pdf'")

        message = MIMEMultipart()
        message["Subject"] = self.subject
        message["From"] = self.sender_email
        message["To"] = recipient_email
        html_part = MIMEText(self.body)
        message.attach(html_part)
        message.attach(part)

        with smtplib.SMTP_SSL("smtp.gmail.com", 465) as server:
            server.login(self.sender_email, self.sender_password)
            server.sendmail(self.sender_email, recipient_email, message.as_string())
