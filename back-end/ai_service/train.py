import numpy as np
import torch
from dataclasses import dataclass
from typing import Any, Dict, List, Union
from datasets import load_dataset, DatasetDict, Audio 
from transformers import (
    WhisperForConditionalGeneration,
    WhisperProcessor,
    Seq2SeqTrainingArguments,
    Seq2SeqTrainer,
)
import evaluate


model_name = "openai/whisper-tiny"
language = "english"
task = "transcribe"

dataset = load_dataset("PolyAI/minds14", "en-US")
dataset = dataset["train"].train_test_split(test_size=0.3)

dataset = dataset.cast_column("audio", Audio(sampling_rate=16000))

model = WhisperForConditionalGeneration.from_pretrained(model_name)
processor = WhisperProcessor.from_pretrained(model_name)


def prepare_dataset(batch):


    processed_dataset = processor(
        batch["audio"]["array"], 
        sampling_rate=16000,
        text=batch["transcription"],
        language="en",
        task=task,

    )

    return processed_dataset

dataset["train"] = dataset["train"].map(
    prepare_dataset,
    remove_columns=dataset["train"].column_names,
    batched=True,
    batch_size=16
)

print("Preparing test dataset (in small batches)...")
dataset["test"] = dataset["test"].map(
    prepare_dataset,
    remove_columns=dataset["test"].column_names,
    batched=True,
    batch_size=16
)


@dataclass
class DataCollator:
    processor: Any

    def __call__(self, features: List[Dict[str, Union[List[int], torch.Tensor]]]) -> Dict[str, torch.Tensor]:
        input_features = [{"input_features": feature["input_features"]} for feature in features]
        label_features = [{"input_ids": feature["labels"]} for feature in features]

        batch = self.processor.feature_extractor.pad(input_features, return_tensors="pt")
        labels_batch = self.processor.tokenizer.pad(label_features, return_tensors="pt")

        labels = labels_batch["input_ids"].masked_fill(labels_batch.attention_mask.ne(1), -100)

        if (labels[:, 0] == self.processor.tokenizer.bos_token_id).all().cpu().item():
            labels = labels[:, 1:]

        batch["labels"] = labels
        return batch

data_collator = DataCollator(processor=processor)



metric = evaluate.load("wer")

def compute_metrics(pred):
    pred_ids = pred.predictions
    label_ids = pred.label_ids

    label_ids[label_ids == -100] = processor.tokenizer.pad_token_id

    pred_str = processor.tokenizer.batch_decode(pred_ids, skip_special_tokens=True)
    label_str = processor.tokenizer.batch_decode(label_ids, skip_special_tokens=True)

    wer = 100 * metric.compute(predictions=pred_str, references=label_str)
    return {"wer": wer}


training_args = Seq2SeqTrainingArguments(
    output_dir="./hoshibatchi-ai",
    per_device_train_batch_size=16,

    eval_strategy="epoch",  
    save_strategy="epoch",  

    num_train_epochs=10,
    learning_rate=1e-5,
    weight_decay=0.01,
    predict_with_generate=True,
    fp16=True,  


    logging_strategy="epoch",
)


trainer = Seq2SeqTrainer(
    args=training_args,
    model=model,
    train_dataset=dataset["train"],
    eval_dataset=dataset["test"],
    data_collator=data_collator,
    compute_metrics=compute_metrics,
    tokenizer=processor.tokenizer
)


trainer.train()


final_model_path = "./hoshibatchi-ai"
trainer.save_model(final_model_path)
processor.save_pretrained(final_model_path)

