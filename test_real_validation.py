import json
import requests
import sys

# Test reading first 100 lines of user's dataset
with open('data/contracts/ledgar_finetune_val.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

# Save truncated version
with open('data/contracts/ledgar_finetune_val_truncated.json', 'w', encoding='utf-8') as f:
    json.dump(data[:50], f)

url = "http://localhost:8080/api/v1/datasets"
files = {'file': open('data/contracts/ledgar_finetune_val_truncated.json', 'rb')}
response = requests.post(url, files=files)
print(f"Validation Response: {response.text}")
