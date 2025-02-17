import librosa
import soundfile
import torch
import os
from transformers import Wav2Vec2ForCTC, Wav2Vec2Processor
from constants import constants


class ASRModel:
    def __init__(self):
        self.model_name = "facebook/wav2vec2-base-960h"
        self.resample_file_path = constants.resample_file_path
        self.sampling_rate = 16000
        self.block_length = 30

    def instantiate_model(self):
        processor = Wav2Vec2Processor.from_pretrained(self.model_name)
        wav_model = Wav2Vec2ForCTC.from_pretrained(self.model_name)
        return wav_model, processor

    def resample_file(self, file: str, file_name: str):
        audio, sound_rate = librosa.load(file, sr=self.sampling_rate)
        resampled_file_path = os.path.join(self.resample_file_path, file_name)
        soundfile.write(resampled_file_path, audio, self.sampling_rate)
        return resampled_file_path

    def asr_transcript(self, processor, model, resampled_path):
        try:
            transcript = ""
            stream = librosa.stream(
                resampled_path,
                block_length=self.block_length,
                frame_length=self.sampling_rate,
                hop_length=self.sampling_rate,
            )

            for n, speech in enumerate(stream):
                separator = " "
                if n % 2 == 0:
                    separator = "\n"
                transcript += (
                    self.generate_transcription(
                        speech=speech, processor=processor, model=model
                    )
                    + separator
                )
            return transcript.strip()

        except Exception as e:
            raise RuntimeError(
                f"Failed to process audio file {resampled_path}: {str(e)}"
            )

    def generate_transcription(self, speech, processor, model):
        if len(speech.shape) > 1:
            speech = speech[:, 0] + speech[:, 1]
        input_values = processor(
            speech, sampling_rate=self.sampling_rate, return_tensors="pt"
        ).input_values
        with torch.no_grad():  # Don't track gradients during inference
            logits = model(input_values).logits
        predicted_ids = torch.argmax(logits, dim=-1)
        transcription = processor.decode(predicted_ids[0])
        return transcription.lower()
