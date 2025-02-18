import librosa
import soundfile
import torch
import os
from transformers import AutoModelForSpeechSeq2Seq, AutoProcessor, pipeline
from constants import constants


class ASRModel:
    def __init__(self):
        self.model_id = "openai/whisper-tiny.en"
        self.device = "cpu"
        self.torch_dtype = torch.float32
        self.resample_file_path = constants.resample_file_path
        self.sampling_rate = 16000
        self.chunk_length = 30

    def instantiate_model(self):
        model = AutoModelForSpeechSeq2Seq.from_pretrained(
            self.model_id,
            torch_dtype=self.torch_dtype,
            low_cpu_mem_usage=True,
            use_safetensors=True,
        )
        model.to(self.device)
        processor = AutoProcessor.from_pretrained(self.model_id)
        return model, processor

    def resample_file(self, file: str, file_name: str):
        audio, sound_rate = librosa.load(file, sr=self.sampling_rate)
        resampled_file_path = os.path.join(self.resample_file_path, file_name)
        soundfile.write(resampled_file_path, audio, self.sampling_rate)
        return resampled_file_path

    def generate_transcript(self, model, processor, file):
        try:
            pipe = pipeline(
                task="automatic-speech-recognition",
                model=model,
                tokenizer=processor.tokenizer,
                feature_extractor=processor.feature_extractor,
                torch_dtype=self.torch_dtype,
                device=self.device,
                chunk_length_s=self.chunk_length,
            )

            result = pipe(file)
            return result["text"]
        except Exception as e:
            raise RuntimeError(f"failed to process audi file: {str(e)}")
