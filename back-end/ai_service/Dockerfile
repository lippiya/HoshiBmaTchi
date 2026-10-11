FROM python:3.10-slim

RUN apt-get update && apt-get install -y \
    ffmpeg \
    libsndfile1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY ai_service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY ai_service/ .

RUN mkdir -p /app/bin && ln -s /usr/bin/ffmpeg /app/bin/ffmpeg

EXPOSE 5000

CMD ["python", "app.py"]
