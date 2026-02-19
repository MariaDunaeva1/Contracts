"""
Download LEDGAR dataset from HuggingFace and convert to fine-tuning chat format.
Output: ledgar_finetune_train.json and ledgar_finetune_val.json
"""
import json
from datasets import load_dataset
from collections import Counter

print("Downloading LEDGAR dataset from HuggingFace...")
dataset = load_dataset("coastalcph/lex_glue", "ledgar")

# Get label names
label_names = dataset['train'].features['label'].names
print(f"\nLabel names ({len(label_names)} categories):")
for i, name in enumerate(label_names):
    print(f"  {i}: {name}")

def convert_split(split_data, output_file):
    """Convert a dataset split to chat/messages JSON format."""
    entries = []
    for example in split_data:
        label_name = label_names[example['label']]
        entry = {
            "messages": [
                {
                    "role": "user",
                    "content": f"Classify this contract clause:\n\n{example['text']}"
                },
                {
                    "role": "assistant",
                    "content": label_name
                }
            ]
        }
        entries.append(entry)
    
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(entries, f, indent=2, ensure_ascii=False)
    
    # Stats
    labels = [label_names[ex['label']] for ex in split_data]
    counter = Counter(labels)
    print(f"\n--- {output_file} ---")
    print(f"  Total entries: {len(entries)}")
    print(f"  Unique labels: {len(counter)}")
    print(f"  Top 10 labels:")
    for label, count in counter.most_common(10):
        print(f"    {count:>5}x  {label}")
    
    return entries

# Convert train split
print("\nConverting train split...")
train_entries = convert_split(dataset['train'], 'ledgar_finetune_train.json')

# Convert validation split
print("\nConverting validation split...")
val_entries = convert_split(dataset['validation'], 'ledgar_finetune_val.json')

# Convert test split  
print("\nConverting test split...")
test_entries = convert_split(dataset['test'], 'ledgar_finetune_test.json')

# File sizes
import os
for f in ['ledgar_finetune_train.json', 'ledgar_finetune_val.json', 'ledgar_finetune_test.json']:
    size_mb = os.path.getsize(f) / (1024 * 1024)
    print(f"\n{f}: {size_mb:.1f} MB")

print("\n✅ Done! Files ready to upload to Kaggle.")
