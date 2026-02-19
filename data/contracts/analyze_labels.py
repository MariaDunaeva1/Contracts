import json
from collections import Counter

with open('finetune_train.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

labels = []
for entry in data:
    for m in entry['messages']:
        if m['role'] == 'assistant':
            labels.append(m['content'])

counter = Counter(labels)
print(f"Total entries: {len(labels)}")
print(f"Unique labels: {len(counter)}")
print()
print("--- Labels sorted by frequency ---")
for label, count in counter.most_common():
    print(f"  {count:>4}x  {label}")
