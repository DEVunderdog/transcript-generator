import librosa
import soundfile
import torch
import os
from transformers import Wav2Vec2ForCTC, Wav2Vec2Processor


class ASRModel:
    def __init__(self, resample_files_path: str):
        self.model_name = "facebook/wav2vec2-base-960h"
        self.resample_file_path = resample_files_path
        self.sampling_rate = 16000

    def instantiate_model(self):
        processor = Wav2Vec2Processor.from_pretrained(self.model_name)
        model = Wav2Vec2ForCTC.from_pretrained(self.model_name)
        return model, processor

    def resample_file(self, file: str, sr: float):
        audio, sound_rate = librosa.load(file, sr=self.sampling_rate)
        length = librosa.get_duration(audio, sound_rate)
        resampled_file_path = os.path.join(self.resample_file_path, file)
        soundfile.write(resampled_file_path, audio, sound_rate)
        return resampled_file_path, length

    def generate_transcription(self, speech, processor, model):
        if len(speech.shape) > 1:
            speech = speech[:, 0] + speech[:, 1]
        input_values = processor(speech, sampling_rate = self.sampling_rate, return_tensors="pt").input_values
        logits = model(input_values).logits
        predicted_ids = torch.argmax(logits, dim=-1)
        transcription = processor.decode(predicted_ids[0])
        return transcription.lower()

    def asr_transcript(self, processor, model, resampled_path, length, block_length):
        chunks = length//block_length
        if length%block_length != 0:
            chunks += 1
        transcript = ""
        stream = librosa.stream(resampled_path, block_length=block_length, frame_length=self.sampling_rate, hop_length=self.sampling_rate)

        for n, speech in enumerate(stream):
            separator = ' '
            if n % 2 == 0:
                separator = '\n'
            transcript += self.generate_transcription(speech=speech, processor=processor, model=model) + separator
        return transcript
