import os
import subprocess
import torch
import torchaudio
import torchaudio.transforms as T
from transformers import WhisperProcessor, WhisperForConditionalGeneration
import numpy as np
from flask import Flask, request, jsonify
from flask_cors import CORS
from flask_compress import Compress
import logging

app = Flask(__name__)
CORS(app)
Compress(app)

log_level = os.getenv("LOG_LEVEL", "INFO")
logging.basicConfig(level=getattr(logging, log_level))
logger = logging.getLogger(__name__)

model_name = "./whisper-tiny"
processor = WhisperProcessor.from_pretrained(model_name)
model = WhisperForConditionalGeneration.from_pretrained(model_name)

@app.route("/transcribe", methods=["POST"])
def transcribe():
    logger.info("Transcribe request received")
    if "audio" not in request.files:
        logger.error("No audio file provided")
        return jsonify({"error": "No audio file provided"}), 400

    audio_file = request.files["audio"]
    audio_path = "/tmp/audio_input.webm"
    audio_file.save(audio_path)

    wav_audio_path = "/tmp/audio_input.wav"
    subprocess.run(['ffmpeg', '-i', audio_path, '-y', wav_audio_path], check=True)

    audio_input, sample_rate = torchaudio.load(wav_audio_path)

    
    if audio_input.shape[0] > 1:
        audio_input = audio_input.mean(dim=0, keepdim=True)

    
    resampler = T.Resample(orig_freq=sample_rate, new_freq=16000)
    audio_input = resampler(audio_input)

    
    audio_input = audio_input.squeeze(0).numpy()

    
    inputs = processor(audio_input, return_tensors="pt", sampling_rate=16000)
    with torch.no_grad():
        predicted_ids = model.generate(inputs["input_features"])

    transcription = processor.decode(predicted_ids[0], skip_special_tokens=True)

    os.remove(audio_path)
    os.remove(wav_audio_path)

    logger.info(f"Transcription completed successfully")
    return jsonify({"transcription": transcription})

if __name__ == "__main__":
    debug_mode = log_level == "DEBUG"
    app.run(debug=debug_mode, use_reloader=False, host="0.0.0.0", port=5000)